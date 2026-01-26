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
	"context"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed schemas/*.md
var schemas embed.FS

// GetLogSchemaRequest defines the requested log type schema.
type GetLogSchemaRequest struct {
	LogType string `json:"log_type" jsonschema:"The type of log to get schema for. Supported values are: ['k8s_audit_logs', 'k8s_application_logs', 'k8s_event_logs']."`
}

var supportedLogTypes = map[string]bool{
	"k8s_audit_logs":       true,
	"k8s_application_logs": true,
	"k8s_event_logs":       true,
}

func installGetLogSchemas(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_log_schema",
		Description: "Get the schema for a specific log type.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, getLogSchema)
}

func getLogSchema(_ context.Context, _ *mcp.CallToolRequest, req *GetLogSchemaRequest) (*mcp.CallToolResult, any, error) {
	if supportedLogTypes[req.LogType] {
		fileName := fmt.Sprintf("%s.md", req.LogType)
		filePath := filepath.Join("schemas", fileName)
		content, err := schemas.ReadFile(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("could not find schema for log_type %s: %w", req.LogType, err)
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(content)},
			},
		}, nil, nil
	}
	return nil, nil, fmt.Errorf("unsupported log_type: %s", req.LogType)
}
