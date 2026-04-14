// Copyright 2026 Google LLC
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

package manifestgen

import (
	"context"
	"fmt"

	"cloud.google.com/go/vertexai/genai"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Agent handles manifest generation using Vertex AI.
type Agent struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewAgent creates a new Agent.
func NewAgent(ctx context.Context, cfg *config.Config) (*Agent, error) {
	projectID := cfg.DefaultProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("default project ID not set in config")
	}
	location := cfg.DefaultLocation()
	if location == "" {
		location = "us-central1" // Default fallback
	}

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex client: %w", err)
	}

	// Use a default model, e.g., gemini-2.5-pro
	model := client.GenerativeModel("gemini-2.5-pro")

	return &Agent{
		client: client,
		model:  model,
	}, nil
}

// GenerateManifest generates a Kubernetes manifest based on the prompt.
func (a *Agent) GenerateManifest(ctx context.Context, prompt string) (string, error) {
	fullPrompt := fmt.Sprintf("Generate a valid Kubernetes manifest for the following request. Return ONLY the YAML content, no markdown blocks or explanations.\n\nRequest: %s", prompt)

	resp, err := a.model.GenerateContent(ctx, genai.Text(fullPrompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from model")
	}

	var result string
	for _, part := range resp.Candidates[0].Content.Parts {
		if text, ok := part.(genai.Text); ok {
			result += string(text)
		}
	}

	return result, nil
}

// Install registers the tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	agent, err := NewAgent(ctx, c)
	if err != nil {
		return err
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate_manifest",
		Description: "Generates a Kubernetes manifest using Vertex AI based on a description.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args *struct {
		Prompt string `json:"prompt" jsonschema:"The description of the manifest to generate. e.g. 'nginx deployment with 3 replicas'"`
	}) (*mcp.CallToolResult, any, error) {
		manifest, err := agent.GenerateManifest(ctx, args.Prompt)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: manifest,
				},
			},
		}, nil, nil
	})

	return nil
}
