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

// Package manifestgen provides an agent for generating Kubernetes manifests.
package manifestgen

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

//go:embed instruction.md
var instructionTemplate string

// Agent handles manifest generation via ADK.
type Agent struct {
	cfg            *config.Config
	adkRunner      *runner.Runner
	sessionService session.Service
}

// NewAgent creates a new Agent attached to a specific text generator model.
func NewAgent(llm model.LLM, cfg *config.Config) (*Agent, error) {
	if llm == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	sessSvc := session.InMemoryService()

	adkAgent, err := llmagent.New(llmagent.Config{
		Name:        "manifest_agent",
		Description: "Agent specialized in generating and validating Kubernetes manifests.",
		Model:       llm,
		Instruction: instructionTemplate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK agent: %w", err)
	}

	adkRunner, err := runner.New(runner.Config{
		AppName:        "gke-mcp",
		Agent:          adkAgent,
		SessionService: sessSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK runner: %w", err)
	}

	return &Agent{
		cfg:            cfg,
		adkRunner:      adkRunner,
		sessionService: sessSvc,
	}, nil
}

// Run executes the agent using the ADK runner.
func (a *Agent) Run(ctx context.Context, prompt string, sessionID string) (string, error) {
	// Ensure session exists
	_, err := a.sessionService.Get(ctx, &session.GetRequest{
		AppName:   "gke-mcp",
		UserID:    "default-user",
		SessionID: sessionID,
	})
	if err != nil {
		_, err = a.sessionService.Create(ctx, &session.CreateRequest{
			AppName:   "gke-mcp",
			UserID:    "default-user",
			SessionID: sessionID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to create session: %w", err)
		}
	}

	msg := &genai.Content{
		Parts: []*genai.Part{{Text: prompt}},
	}

	events := a.adkRunner.Run(ctx, "default-user", sessionID, msg, agent.RunConfig{})

	var builder strings.Builder
	for event, err := range events {
		if err != nil {
			return "", err
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				builder.WriteString(part.Text)
			}
		}
	}

	return builder.String(), nil
}

// Install registers the tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	// Create a new Gemini model backed by Vertex AI via ADK
	llm, err := gemini.NewModel(ctx, "gemini-2.5-pro", &genai.ClientConfig{
		Project:  c.DefaultProjectID(),
		Backend:  genai.BackendVertexAI,
		Location: c.DefaultLocation(),
	})
	if err != nil {
		return fmt.Errorf("failed to create gemini model: %w", err)
	}

	agent, err := NewAgent(llm, c)
	if err != nil {
		return err
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate_manifest",
		Description: "Generates a Kubernetes manifest using Vertex AI based on a description.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args *struct {
		Prompt    string `json:"prompt" jsonschema:"The description of the manifest to generate. e.g. 'nginx deployment with 3 replicas'"`
		SessionID string `json:"session_id,omitempty" jsonschema:"Optional. A unique identifier to maintain conversation history across multiple tool calls. If not provided, a new random ID will be generated."`
	}) (*mcp.CallToolResult, any, error) {
		sessID := args.SessionID
		if sessID == "" {
			sessID = uuid.New().String()
		}
		manifest, err := agent.Run(ctx, args.Prompt, sessID)
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
