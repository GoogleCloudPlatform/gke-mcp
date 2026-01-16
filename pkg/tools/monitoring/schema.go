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
	"embed"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schemas/*.md
var metricSchemas embed.FS

type GetMetricSchemaRequest struct {
	ResourceType string `json:"resource_type" jsonschema:"The monitored resource type to get schema for. Supported values are: ['k8s_cluster', 'k8s_container', 'k8s_control_plane_component', 'k8s_entity', 'k8s_node', 'k8s_node_pool', 'k8s_pod', 'k8s_scale', 'k8s_service']."`
}

var supportedResourceTypes = map[string]bool{
	"k8s_cluster":                 true,
	"k8s_container":               true,
	"k8s_control_plane_component": true,
	"k8s_entity":                  true,
	"k8s_node":                    true,
	"k8s_node_pool":               true,
	"k8s_pod":                     true,
	"k8s_scale":                   true,
	"k8s_service":                 true,
}

func installGetMetricSchemas(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_metric_schema",
		Description: "Get the schema for a specific GKE monitored resource type used in metric queries.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, getMetricSchema)
}

func getMetricSchema(_ context.Context, _ *mcp.CallToolRequest, req *GetMetricSchemaRequest) (*mcp.CallToolResult, any, error) {
	if !supportedResourceTypes[req.ResourceType] {
		return nil, nil, fmt.Errorf("unsupported resource_type: %s", req.ResourceType)
	}

	fileName := fmt.Sprintf("%s.md", req.ResourceType)
	filePath := filepath.Join("schemas", fileName)
	content, err := metricSchemas.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("could not find schema for resource_type %s: %w", req.ResourceType, err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(content)},
		},
	}, nil, nil
}
