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

// Package monitoring provides tools for GKE-related monitoring data.
package monitoring

import (
	"context"
	"fmt"
	"log"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

type handlers struct {
	c *config.Config
}

type listMonitoredResourceDescriptorsArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
}

// Install registers monitoring tools with the MCP server.
func Install(_ context.Context, s *mcp.Server, c *config.Config) error {
	h := &handlers{
		c: c,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_monitored_resource_descriptors",
		Description: "List monitored resource descriptors(schema) related to GKE for this project. Prefer to use this tool instead of gcloud",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.listMRDescriptor)

	installGetMetricSchemas(s)

	return nil
}

func (h *handlers) listMRDescriptor(ctx context.Context, _ *mcp.CallToolRequest, args *listMonitoredResourceDescriptorsArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.ProjectID == "" {
		return nil, nil, fmt.Errorf("project_id argument cannot be empty")
	}
	c, err := monitoring.NewMetricClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("Failed to close monitoring client: %v\n", err)
		}
	}()
	req := &monitoringpb.ListMonitoredResourceDescriptorsRequest{
		Name: fmt.Sprintf("projects/%s", args.ProjectID),
	}
	it := c.ListMonitoredResourceDescriptors(ctx, req)
	builder := new(strings.Builder)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		builder.WriteString(protojson.Format(resp))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: builder.String()},
		},
	}, nil, nil
}
