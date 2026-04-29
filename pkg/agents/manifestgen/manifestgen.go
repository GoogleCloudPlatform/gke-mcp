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
	"log"
	"os"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/giq"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
	"google.golang.org/genai"
)

//go:embed instruction.md
var instructionTemplate string

const defaultModel = "gemini-2.5-pro"

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

	giqTool, err := functiontool.New(
		functiontool.Config{
			Name:        "giq_generate_manifest",
			Description: "Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud",
		},
		func(ctx tool.Context, args giq.GenerateInferenceManifestArgs) (string, error) {
			return giq.GenerateInferenceManifest(ctx, &args)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create giq tool: %w", err)
	}

	fetchModelsTool, err := functiontool.New(
		functiontool.Config{
			Name:        "giq_fetch_models",
			Description: "List all AI models available for GKE via GKE Inference Quickstart (GIQ). Open-source models follow the Huggingface Hub `owner/model_name` format.",
		},
		func(ctx tool.Context, _ struct{}) (string, error) {
			return giq.FetchModels(ctx)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create giq fetch models tool: %w", err)
	}

	adkAgent, err := llmagent.New(llmagent.Config{
		Name:        "manifest_agent",
		Description: "Agent specialized in generating and validating Kubernetes manifests.",
		Model:       llm,
		Instruction: instructionTemplate,
		Tools:       []tool.Tool{giqTool, fetchModelsTool},
		BeforeModelCallbacks: []llmagent.BeforeModelCallback{
			func(ctx agent.CallbackContext, llmRequest *model.LLMRequest) (*model.LLMResponse, error) {
				// Inject user content if Contents is empty to avoid content loss.
				if len(llmRequest.Contents) == 0 {
					userContent := ctx.UserContent()
					if userContent != nil {
						userContent.Role = "user"
						llmRequest.Contents = append(llmRequest.Contents, userContent)
					}
				}

				if os.Getenv("GKE_MCP_DEBUG") == "true" {
					log.Printf("--- Before Model Call ---")
					log.Printf("Model: %s", llmRequest.Model)
					if llmRequest.Config != nil {
						log.Printf("Config: %+v", llmRequest.Config)
					}
					log.Printf("Contents count: %d", len(llmRequest.Contents))
					for i, c := range llmRequest.Contents {
						log.Printf("Content %d (Role: %s):", i, c.Role)
						for j, p := range c.Parts {
							log.Printf("  Part %d: %q", j, p.Text)
						}
					}
				}
				return nil, nil
			},
		},
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
		Role:  "user",
		Parts: []*genai.Part{{Text: prompt}},
	}

	events := a.adkRunner.Run(ctx, "default-user", sessionID, msg, agent.RunConfig{})

	var builder strings.Builder
	if os.Getenv("GKE_MCP_DEBUG") == "true" {
		log.Printf("=== New Run with prompt: %q ===", prompt)
	}

	for event, err := range events {
		if err != nil {
			if os.Getenv("GKE_MCP_DEBUG") == "true" {
				log.Printf("Error event: %v", err)
			}
			return "", err
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if os.Getenv("GKE_MCP_DEBUG") == "true" {
					log.Printf("Model Part: %q", part.Text)
				}
				builder.WriteString(part.Text)
			}
		}
	}

	if os.Getenv("GKE_MCP_DEBUG") == "true" {
		log.Printf("Final result: %q", builder.String())
	}

	return builder.String(), nil
}

// Install registers the tool with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	// Create a new Gemini model backed by Vertex AI via ADK
	llm, err := gemini.NewModel(ctx, defaultModel, &genai.ClientConfig{
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
