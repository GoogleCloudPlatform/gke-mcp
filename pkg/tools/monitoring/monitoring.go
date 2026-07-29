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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/registry"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/iterator"
	monitoringv1 "google.golang.org/api/monitoring/v1"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

type handlers struct {
	c        *config.Config
	endpoint string
}

type listMonitoredResourceDescriptorsArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
}

type queryPrometheusArgs struct {
	ProjectID string `json:"project_id" jsonschema:"Required. GCP project ID."`
	Query     string `json:"query" jsonschema:"Required. PromQL query string."`
	Time      string `json:"time,omitempty" jsonschema:"Optional. Evaluation time (RFC3339 or unix timestamp)."`
	Timeout   string `json:"timeout,omitempty" jsonschema:"Optional. Query timeout (e.g., '10s')."`
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

	registry.RegisterTool(s, c, &mcp.Tool{
		Name:        "query_prometheus",
		Description: "Query Cloud Monitoring metrics using PromQL (instant query). Returns Prometheus-compatible JSON response.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.queryPrometheus)

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

func (h *handlers) queryPrometheus(ctx context.Context, _ *mcp.CallToolRequest, args *queryPrometheusArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		return nil, nil, fmt.Errorf("project_id argument cannot be empty")
	}
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query argument cannot be empty")
	}

	opts := []option.ClientOption{
		option.WithUserAgent(h.c.UserAgent()),
	}
	if h.endpoint != "" {
		opts = append(opts, option.WithEndpoint(h.endpoint), option.WithoutAuthentication())
	}

	svc, err := monitoringv1.NewService(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create monitoring service: %w", err)
	}

	req := &monitoringv1.QueryInstantRequest{
		Query:   args.Query,
		Time:    args.Time,
		Timeout: args.Timeout,
	}

	resp, err := svc.Projects.Location.Prometheus.Api.V1.Query("projects/"+args.ProjectID, "global", req).Context(ctx).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query prometheus metrics: %w", err)
	}

	if resp != nil && resp.Data != nil {
		if str, ok := resp.Data.(string); ok {
			decoded, err := base64.StdEncoding.DecodeString(str)
			if err == nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: string(decoded)},
					},
				}, nil, nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: str},
				},
			}, nil, nil
		}
		bytes, err := json.Marshal(resp.Data)
		if err == nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(bytes)},
				},
			}, nil, nil
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "{}"},
		},
	}, nil, nil
}
