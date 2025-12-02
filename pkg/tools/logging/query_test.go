// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"encoding/json"
	"testing"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/google/go-cmp/cmp"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLogQueryRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     LogQueryRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Query:     "severity=ERROR",
			},
			wantErr: false,
		},
		{
			name:    "missing project id",
			req:     LogQueryRequest{},
			wantErr: true,
		},
		{
			name: "limit too high",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Limit:     101,
			},
			wantErr: true,
		},
		{
			name: "invalid since duration",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Since:     "invalid",
			},
			wantErr: true,
		},
		{
			name: "since and time_range both set",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Since:     "1h",
				TimeRange: &TimeRange{
					StartTime: time.Now(),
				},
			},
			wantErr: true,
		},
		{
			name: "invalid format template",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Format:    "{{.invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.validate(); (err != nil) != tt.wantErr {
				t.Errorf("LogQueryRequest.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildListLogEntriesRequest(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		req  LogQueryRequest
		want *loggingpb.ListLogEntriesRequest
	}{
		{
			name: "basic request",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Query:     "severity=ERROR",
				Limit:     10,
			},
			want: &loggingpb.ListLogEntriesRequest{
				ResourceNames: []string{"projects/test-project"},
				Filter:        "severity=ERROR",
				PageSize:      10,
				OrderBy:       "timestamp asc",
			},
		},
		{
			name: "request with time range",
			req: LogQueryRequest{
				ProjectID: "test-project",
				Query:     "severity=ERROR",
				Limit:     10,
				TimeRange: &TimeRange{
					StartTime: now.Add(-1 * time.Hour),
					EndTime:   now,
				},
			},
			want: &loggingpb.ListLogEntriesRequest{
				ResourceNames: []string{"projects/test-project"},
				Filter:        `severity=ERROR AND timestamp >= "` + now.Add(-1*time.Hour).Format(time.RFC3339) + `" AND timestamp <= "` + now.Format(time.RFC3339) + `"`,
				PageSize:      10,
				OrderBy:       "timestamp asc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildListLogEntriesRequest(&tt.req)
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("buildListLogEntriesRequest() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatter(t *testing.T) {
	entry := &loggingpb.LogEntry{
		Payload: &loggingpb.LogEntry_TextPayload{
			TextPayload: "test log",
		},
		Severity:  ltype.LogSeverity_ERROR,
		Timestamp: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	jsonEntry := &loggingpb.LogEntry{
		Payload: &loggingpb.LogEntry_JsonPayload{
			JsonPayload: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"key": structpb.NewStringValue("value"),
				},
			},
		},
		Severity:  ltype.LogSeverity_ERROR,
		Timestamp: timestamppb.New(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)),
	}

	tests := []struct {
		name    string
		req     LogQueryRequest
		entry   *loggingpb.LogEntry
		want    string
		wantErr bool
		isJSON  bool
	}{
		{
			name:  "json formatter text payload",
			req:   LogQueryRequest{},
			entry: entry,
			want: `{
  "severity": "ERROR",
  "textPayload": "test log",
  "timestamp": "2023-01-01T00:00:00Z"
}`,
			wantErr: false,
			isJSON:  true,
		},
		{
			name:  "json formatter json payload",
			req:   LogQueryRequest{},
			entry: jsonEntry,
			want: `{
  "jsonPayload": {
    "key": "value"
  },
  "severity": "ERROR",
  "timestamp": "2023-01-01T00:00:00Z"
}`,
			wantErr: false,
			isJSON:  true,
		},
		{
			name: "template formatter",
			req: LogQueryRequest{
				Format: "{{.textPayload}} - {{.severity}}",
			},
			entry:   entry,
			want:    "test log - ERROR",
			wantErr: false,
			isJSON:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := formatterForRequest(&tt.req)
			if err != nil {
				t.Fatalf("formatterForRequest() error = %v", err)
			}
			got, err := f.format(tt.entry)
			if (err != nil) != tt.wantErr {
				t.Errorf("formatter.format() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.isJSON {
				var gotMap, wantMap map[string]interface{}
				if err := json.Unmarshal([]byte(got), &gotMap); err != nil {
					t.Fatalf("failed to unmarshal got JSON: %v", err)
				}
				if err := json.Unmarshal([]byte(tt.want), &wantMap); err != nil {
					t.Fatalf("failed to unmarshal want JSON: %v", err)
				}
				if diff := cmp.Diff(wantMap, gotMap); diff != "" {
					t.Errorf("formatter.format() mismatch (-want +got):\n%s", diff)
				}
			} else {
				if got != tt.want {
					t.Errorf("formatter.format() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
