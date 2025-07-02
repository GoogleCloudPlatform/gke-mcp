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

	listRecommendationsTool := mcp.NewTool("list_recommendation",
		mcp.WithDescription("List recommendations for GKE. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("project_id", mcp.DefaultString(c.DefaultProjectID()), mcp.Description("GCP project ID. Use the default if the user doesn't provide it.")),
		mcp.WithString("location", mcp.Description("GKE cluster location. Leave this empty if the user doesn't doesn't provide it.")),
	)
	s.AddTool(listRecommendationsTool, h.listRecommendations)
}

func (h *handlers) listRecommendations(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", h.c.DefaultProjectID())
	if projectID == "" {
		return mcp.NewToolResultError("project_id argument not set"), nil
	}
	location, _ := request.RequireString("location")
	if location == "" {
		location = "-"
	}
  
	c, err := recommender.NewClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()

	req :=  &recommenderpb.ListRecommendationsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/recommender/google.container.DiagnosisRecommender", projectID, location),
	}
	resp := c.ListRecommendations(ctx, req)
	return mcp.NewToolResultText(protojson.Format(resp)), nil
}
