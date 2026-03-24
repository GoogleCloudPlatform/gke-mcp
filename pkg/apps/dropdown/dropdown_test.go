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

package dropdown

import (
	"context"
	"reflect"
	"testing"
)

func TestDropdownHandler(t *testing.T) {
	tests := []struct {
		name    string
		args    *dropdownArgs
		want    PendingResponse
		wantErr bool
	}{
		{
			name: "valid options",
			args: &dropdownArgs{
				Title:   "Select a cluster",
				Options: []string{"cluster1", "cluster2"},
			},
			want: PendingResponse{
				Status:  "PENDING_USER_INPUT",
				Options: []string{"cluster1", "cluster2"},
				Message: "Present these options to the user. Wait until selection is made",
			},
			wantErr: false,
		},
		{
			name: "nil options",
			args: &dropdownArgs{
				Title:   "Select a cluster",
				Options: nil,
			},
			want: PendingResponse{
				Status:  "PENDING_USER_INPUT",
				Options: nil,
				Message: "Present these options to the user. Wait until selection is made",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := dropdownHandler(context.Background(), nil, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("dropdownHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatalf("dropdownHandler() returned nil result")
				}
				payload, ok := result.StructuredContent.(PendingResponse)
				if !ok {
					t.Fatalf("expected StructuredContent to be type PendingResponse, got %T", result.StructuredContent)
				}
				if !reflect.DeepEqual(payload, tt.want) {
					t.Errorf("dropdownHandler() payload = %v, want %v", payload, tt.want)
				}
			}
		})
	}
}
