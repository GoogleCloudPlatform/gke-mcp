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
	"strings"

	logging "cloud.google.com/go/logging/apiv2"
	loggingpb "cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

type handlers struct {
	c *config.Config
}

func Install(s *server.MCPServer, c *config.Config) {
	h := &handlers{
		c: c,
	}

	listMRSchemaTool := mcp.NewTool("list_mr_schema",
		mcp.WithDescription("List monitored resource descriptors(Schema) for this project. It should contains the schema for logs and metrics etc. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listMRSchemaTool, h.listMRSchema)
}

func (h *handlers) listMRSchema(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c, err := logging.NewClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()
	req := &loggingpb.ListMonitoredResourceDescriptorsRequest{}
	it := c.ListMonitoredResourceDescriptors(ctx, req)
	builder := new(strings.Builder)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		builder.WriteString(protojson.Format(resp))
	}
	return mcp.NewToolResultText(builder.String()), nil
}
