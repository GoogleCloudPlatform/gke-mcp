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

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

//go:embed schemas/*.md
var schemas embed.FS

type GetLogSchemaRequest struct {
	LogType string `json:"log_type"`
}

func installGetLogSchemas(s *server.MCPServer) {
	getLogSchemaTool := mcp.NewTool("get_log_schema",
		mcp.WithDescription("Get the schema for a specific log type."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("log_type", mcp.Description("The type of log to get schema for. Supported values are: ['k8s_audit_logs', 'k8s_application_logs']."), mcp.Required()),
	)
	s.AddTool(getLogSchemaTool, mcp.NewTypedToolHandler(getLogSchema))
}

func getLogSchema(_ context.Context, _ mcp.CallToolRequest, req GetLogSchemaRequest) (*mcp.CallToolResult, error) {
	switch req.LogType {
	case "k8s_audit_logs":
		fileName := fmt.Sprintf("%s.md", req.LogType)
		filePath := filepath.Join("schemas", fileName)
		content, err := schemas.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not find schema for log_type %s: %v", req.LogType, err)
		}
		return mcp.NewToolResultText(string(content)), nil
	case "k8s_application_logs":
		fileName := fmt.Sprintf("%s.md", req.LogType)
		filePath := filepath.Join("schemas", fileName)
		content, err := schemas.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("could not find schema for log_type %s: %v", req.LogType, err)
		}
		return mcp.NewToolResultText(string(content)), nil
	default:
		return mcp.NewToolResultErrorf("unsupported log_type: %s", req.LogType), nil
	}
}
