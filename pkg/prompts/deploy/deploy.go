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

package deploy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	tooldeploy "github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/deploy"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var gkeDeployTmpl = template.Must(template.New("gke-deploy").Parse(tooldeploy.PromptTemplate))

func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	s.AddPrompt(&mcp.Prompt{
		Name:        "gke:deploy",
		Description: "Deploys a workload to a GKE cluster using a configuration file.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "user_request",
				Description: "A natural language request specifying the configuration file to deploy. e.g. 'my-app.yaml to staging'",
				Required:    true,
			},
		},
	}, gkeDeployHandler)

	return nil
}

func gkeDeployHandler(ctx context.Context, request *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	userRequest := request.Params.Arguments["user_request"]
	if strings.TrimSpace(userRequest) == "" {
		return nil, fmt.Errorf("argument 'user_request' cannot be empty")
	}

	var buf bytes.Buffer
	if err := gkeDeployTmpl.Execute(&buf, map[string]string{"user_request": userRequest}); err != nil {
		return nil, fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return &mcp.GetPromptResult{
		Description: "GKE Deploy Prompt",
		Messages: []*mcp.PromptMessage{
			{
				Content: &mcp.TextContent{
					Text: buf.String(),
				},
				Role: "user",
			},
		},
	}, nil
}
