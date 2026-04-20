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
	"iter"
	"log"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

//go:embed instruction.md
var instructionTemplate string

// GenerativeModel interface defines mockable text generation capabilities for the new SDK.
type GenerativeModel interface {
	GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
}

// Agent handles manifest generation via ADK.
type Agent struct {
	model          GenerativeModel
	cfg            *config.Config
	adkAgent       agent.Agent
	adkRunner      *runner.Runner
	sessionService session.Service
}

// NewAgent creates a new Agent attached to a specific text generator model.
func NewAgent(model GenerativeModel, cfg *config.Config) (*Agent, error) {
	if model == nil {
		return nil, fmt.Errorf("model cannot be nil")
	}

	sessSvc := session.InMemoryService()
	a := &Agent{
		model:          model,
		cfg:            cfg,
		sessionService: sessSvc,
	}

	adkAgent, err := agent.New(agent.Config{
		Name:        "manifest_agent",
		Description: "Agent specialized in generating and validating Kubernetes manifests.",
		Run:         a.adkRun,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK agent: %w", err)
	}
	a.adkAgent = adkAgent

	adkRunner, err := runner.New(runner.Config{
		AppName:        "gke-mcp",
		Agent:          adkAgent,
		SessionService: sessSvc,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create ADK runner: %w", err)
	}
	a.adkRunner = adkRunner

	return a, nil
}

// adkRun handles the agent invocation in ADK framework.
func (a *Agent) adkRun(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
	return func(yield func(*session.Event, error) bool) {
		systemInstruction := instructionTemplate
		
		userPrompt := ""
		if ctx.UserContent() != nil && len(ctx.UserContent().Parts) > 0 {
			userPrompt = ctx.UserContent().Parts[0].Text
		}

		config := &genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(systemInstruction, ""),
		}

		resp, err := a.model.GenerateContent(ctx, "gemini-2.5-pro", []*genai.Content{genai.NewContentFromText(userPrompt, "")}, config)
		if err != nil {
			log.Printf("ERROR in adkRun GenerateContent: %v", err)
			yield(nil, fmt.Errorf("failed to generate content: %w", err))
			return
		}

		event := session.NewEvent(ctx.InvocationID())
		
		var resultText string
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			for _, part := range resp.Candidates[0].Content.Parts {
				resultText += part.Text
			}
		}
		
		event.Content = &genai.Content{
			Parts: []*genai.Part{{Text: resultText}},
		}

		yield(event, nil)
	}
}

// Run executes the agent using the ADK runner.
func (a *Agent) Run(ctx context.Context, prompt string) (string, error) {
	sessionID := "default-session"
	
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

	var result string
	for event, err := range events {
		if err != nil {
			return "", err
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				result += part.Text
			}
		}
	}

	return result, nil
}

// Install registers the tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	// Create a new GenAI client from the new SDK
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  c.DefaultProjectID(),
		Backend:  genai.BackendVertexAI,
		Location: c.DefaultLocation(),
	})
	if err != nil {
		return fmt.Errorf("failed to create genai client: %w", err)
	}

	agent, err := NewAgent(client.Models, c)
	if err != nil {
		return err
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "generate_manifest",
		Description: "Generates a Kubernetes manifest using Vertex AI based on a description.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args *struct {
		Prompt string `json:"prompt" jsonschema:"The description of the manifest to generate. e.g. 'nginx deployment with 3 replicas'"`
	}) (*mcp.CallToolResult, any, error) {
		manifest, err := agent.Run(ctx, args.Prompt)
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
