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

package monitoring

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGetMetricSchema(t *testing.T) {
	tests := []struct {
		name    string
		req     GetMetricSchemaRequest
		wantErr bool
	}{
		{
			name: "valid resource type",
			req: GetMetricSchemaRequest{
				ResourceType: "k8s_cluster",
			},
			wantErr: false,
		},
		{
			name: "invalid resource type",
			req: GetMetricSchemaRequest{
				ResourceType: "invalid_resource_type",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := getMetricSchema(context.Background(), &mcp.CallToolRequest{}, &tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMetricSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
