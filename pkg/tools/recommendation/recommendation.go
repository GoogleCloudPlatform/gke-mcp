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

package recommendation

import (
	"context"
	"fmt"

	recommender "cloud.google.com/go/recommender/apiv1"
	recommenderpb "cloud.google.com/go/recommender/apiv1/recommenderpb"
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

	listRecommendationsTool := mcp.NewTool("list_recommendations",
		mcp.WithDescription("List recommendations for GKE. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("project_id", mcp.DefaultString(c.DefaultProjectID()), mcp.Description("GCP project ID. If not provided, defaults to the GCP project configured in gcloud, if any")),
		mcp.WithString("location", mcp.Required(), mcp.Description("GKE cluster location. This is required by the recommender API")),
	)
	s.AddTool(listRecommendationsTool, h.listProjectRecommendations)
}

func (h *handlers) listProjectRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", h.c.DefaultProjectID())
	if projectID == "" {
		return mcp.NewToolResultError("project_id argument not set"), nil
	}
	location, _ := request.RequireString("location")
	if location == "" {
		return mcp.NewToolResultError("location argument not set"), nil
	}
	c, err := recommender.NewClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()

	req := &recommenderpb.ListRecommendationsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/recommenders/google.container.DiagnosisRecommender", projectID, location),
	}
	it := c.ListRecommendations(ctx, req)
	builder := new(strings.Builder)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		builder.writeString(protojson.Format(resp))
	}
	return mcp.NewToolResultText(builder.String()), nil
}
