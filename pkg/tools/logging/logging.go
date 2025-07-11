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
	"fmt"
	"strings"
	"time"

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

	listLogsTool := mcp.NewTool("list_logs",
		mcp.WithDescription("List all cloud loggings logs for one given GKE cluster in a location in past 24 hours. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_id", mcp.DefaultString(c.DefaultProjectID()), mcp.Description("GCP project ID. If not provided, defaults to the GCP project configured in gcloud, if any")),
		mcp.WithString("location", mcp.Required(), mcp.Description("GKE cluster location. This is required for filtering on cluster")),
		mcp.WithString("cluster_name", mcp.Required(), mcp.Description("GKE cluster name. This is required for filtering on cluster")),
	)
	s.AddTool(listLogsTool, h.listLogs)
}

func (h *handlers) listLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", h.c.DefaultProjectID())
	if projectID == "" {
		return mcp.NewToolResultError("project_id argument not set"), nil
	}
	location, _ := request.RequireString("location")
	if location == "" {
		return mcp.NewToolResultError("location argument not set"), nil
	}
	clusterName, _ := request.RequireString("cluster_name")
	if clusterName == "" {
		return mcp.NewToolResultError("cluster_name argument not set"), nil
	}
	c, err := logging.NewClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()
	currentTime := time.Now()
	// In the first iteration, we start with one day ago.
	// Time based filtering could be included in the future update.
	oneHourAgo := currentTime.Add(-24 * time.Hour)
	filter := fmt.Sprintf(`%s AND timestamp > "%s"`, filterForCluster(clusterName, location), oneHourAgo.Format(time.RFC3339))
	req := &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + projectID},
		Filter:        filter,
		// PageSize is default to be 100k, pagination could be supported in future update.
		PageSize: 100000,
	}
	it := c.ListLogEntries(ctx, req)
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

// buildFilter converts a set of params into a query string
// that can be used as a filter for logs, metrics, and metadata.
func buildFilter(params map[string]string) string {
	var l []string
	for k, v := range params {
		l = append(l, fmt.Sprintf("%s = %q", k, v))
	}

	return strings.Join(l, " AND ")
}

// filterForCluster returns a OnePlatform filter string for cluster that can be used to
// discover telemetry such as logs, metadata, and metrics
func filterForCluster(name string, location string) string {
	return buildFilter(map[string]string{
		"resource.labels.location":     location,
		"resource.labels.cluster_name": name,
	})
}
