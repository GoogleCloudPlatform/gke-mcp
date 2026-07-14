// Copyright 2026 Google LLC
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

package trace

import (
	"testing"
	"time"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

func TestQueryTracesArgs_Validate(t *testing.T) {
	c := config.NewTestConfig("default-project", "us-central1", "test-provider", "test-model")

	tests := []struct {
		name    string
		args    queryTracesArgs
		wantErr bool
	}{
		{
			name: "valid request with all fields",
			args: queryTracesArgs{
				ProjectID: "my-project",
				Since:     "2h",
				Limit:     50,
				Filter:    "+root:http",
			},
			wantErr: false,
		},
		{
			name:    "missing project id returns error",
			args:    queryTracesArgs{},
			wantErr: true,
		},
		{
			name: "limit too high",
			args: queryTracesArgs{
				ProjectID: "my-project",
				Limit:     200,
			},
			wantErr: true,
		},
		{
			name: "invalid since duration",
			args: queryTracesArgs{
				ProjectID: "my-project",
				Since:     "invalid_time",
			},
			wantErr: true,
		},
		{
			name: "negative since duration",
			args: queryTracesArgs{
				ProjectID: "my-project",
				Since:     "-1h",
			},
			wantErr: true,
		},
		{
			name: "start time after end time",
			args: queryTracesArgs{
				ProjectID: "my-project",
				TimeRange: TimeRange{
					StartTime: time.Now().Add(1 * time.Hour),
					EndTime:   time.Now(),
				},
				Since: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.setDefaults(c)
			err := tt.args.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueryTracesArgs_SetDefaults(t *testing.T) {
	c := config.NewTestConfig("expected-default", "us-central1", "test-provider", "test-model")

	args := queryTracesArgs{}
	args.setDefaults(c)
	if args.Limit != 10 {
		t.Errorf("expected Limit to default to 10, got %d", args.Limit)
	}
	if args.MaxSpans != 5 {
		t.Errorf("expected MaxSpans to default to 5, got %d", args.MaxSpans)
	}
	if args.Since != "1h" {
		t.Errorf("expected Since to default to 1h, got %s", args.Since)
	}
}
