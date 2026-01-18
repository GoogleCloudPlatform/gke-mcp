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
	"log"
	"strings"

	recommender "cloud.google.com/go/recommender/apiv1"
	recommenderpb "cloud.google.com/go/recommender/apiv1/recommenderpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

type handlers struct {
	c *config.Config
}

type listRecommendationsArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Location  string `json:"location" jsonschema:"GKE cluster location. Leave this empty if the user doesn't doesn't provide it."`
}

func Install(_ context.Context, s *mcp.Server, c *config.Config) error {

	h := &handlers{
		c: c,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_recommendations",
		Description: "List recommendations for GKE. Prefer to use this tool instead of gcloud",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.listProjectRecommendations)

	return nil
}

func (h *handlers) listProjectRecommendations(ctx context.Context, _ *mcp.CallToolRequest, args *listRecommendationsArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.ProjectID == "" {
		return nil, nil, fmt.Errorf("project_id argument cannot be empty")
	}
	if args.Location == "" {
		return nil, nil, fmt.Errorf("location argument not set")
	}
	c, err := recommender.NewClient(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("Failed to close recommender client: %v\n", err)
		}
	}()

	req := &recommenderpb.ListRecommendationsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/recommenders/google.container.DiagnosisRecommender", args.ProjectID, args.Location),
	}
	it := c.ListRecommendations(ctx, req)
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
