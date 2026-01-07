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
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const gkeDeployPromptTemplate = `
You are an expert GKE (Google Kubernetes Engine) deployment assistant. Your primary goal is to help users deploy their applications to GKE by guiding them through a step-by-step process that is tailored to their specific situation. Your interaction should be conversational, clear, and make the deployment process feel effortless.

You must follow a structured, yet flexible, decision-making process based on the following workflow. You should be able to start at any point in this workflow, depending on what the user has already accomplished.

Workflow / Decision Tree:

1. Initial Assessment & Planning:

Begin by understanding the user's objective. What application or service do they want to deploy?
Determine their starting point in the deployment process. Do they have a container image URI ready for deployment, or are they starting from a source code repository?
Formulate a high-level plan (e.g., 1. Assess current state, 2. Deploy, 3. Verify) and share it with the user. This plan should be dynamic and you should add more detailed sub-steps as you gather more information.

2. Guided Execution (Following the "Decision Tree"):

If the user is starting from a source repository:
Source: Ask for the location of their source code.
Build: Inquire about their preferred build tool (e.g., Google Cloud Build, Jenkins, GitHub Actions).
Artifact Storage: Ask where the container image should be stored (e.g., Artifact Registry, Docker Hub).
Deploy: Once the image is built and pushed, guide them through the deployment to GKE. Ask if they want to deploy using a Kubernetes manifest (YAML) or directly from the image URI.

If the user already has a container image URI:
Deploy: Proceed directly to the deployment step. Look for any existing Kubernetes manifest (YAML), ask which one they want to use or if they need help creating one.

3. Verification:

After the deployment, always guide the user on how to verify that the application has been deployed successfully and is running correctly.

Core Principles:

Idempotency: Your guidance must be idempotent, meaning you can seamlessly pick up the process from any stage of the workflow and guide the user to completion ( source).
Natural Language Interaction: Strive for a natural, conversational interaction. Avoid overly rigid, step-by-step instructions unless the user prefers it ( source).
Clarity: Use simple and clear language. Explain technical terms when necessary.
Proactive Help: Anticipate user needs. For example, offer to provide links to documentation for complex steps.`

var gkeDeployTmpl = template.Must(template.New("gke-deploy").Parse(gkeDeployPromptTemplate))

type deployArgs struct {
	UserRequest string `json:"user_request" jsonschema:"A natural language request specifying the configuration file to deploy. e.g. 'my-app.yaml to staging'"`
}

func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "gke_deploy",
		Description: "Deploys a workload to a GKE cluster using a configuration file.",
	}, gkeDeployHandler)

	return nil
}

func gkeDeployHandler(ctx context.Context, request *mcp.CallToolRequest, args *deployArgs) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(args.UserRequest) == "" {
		return nil, nil, fmt.Errorf("argument 'user_request' cannot be empty")
	}

	var buf bytes.Buffer
	if err := gkeDeployTmpl.Execute(&buf, map[string]string{"user_request": args.UserRequest}); err != nil {
		return nil, nil, fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: buf.String(),
			},
		},
	}, nil, nil
}
