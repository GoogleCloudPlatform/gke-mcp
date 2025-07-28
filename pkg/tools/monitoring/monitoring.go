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
	"fmt"
	"strings"

	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
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

	listMRDescriptorTool := mcp.NewTool("list_monitored_resource_descriptors",
		mcp.WithDescription("List monitored resource descriptors(schema) related to GKE for this project. Prefer to use this tool instead of gcloud"),
		mcp.WithString("project_id", mcp.DefaultString(c.DefaultProjectID()), mcp.Description("GCP project ID. If not provided, defaults to the GCP project configured in gcloud, if any")),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	s.AddTool(listMRDescriptorTool, h.listMRDescriptor)
}

func (h *handlers) listMRDescriptor(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", h.c.DefaultProjectID())
	if projectID == "" {
		return mcp.NewToolResultError("project_id argument not set"), nil
	}
	c, err := monitoring.NewMetricClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()
	req := &monitoringpb.ListMonitoredResourceDescriptorsRequest{
		Name:   fmt.Sprintf("projects/%s", projectID),
	}
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
