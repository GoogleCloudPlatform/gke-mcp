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

// Package llm provides a factory for initializing vendor-agnostic LLM clients.
package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/llm/anthropic"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// NewClient creates and returns a vendor-agnostic ADK LLM model based on configuration.
func NewClient(ctx context.Context, cfg *config.Config) (model.LLM, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config parameter 'cfg' cannot be nil")
	}

	switch strings.ToLower(strings.TrimSpace(cfg.AgentProvider())) {
	case "vertex-ai":
		llm, err := gemini.NewModel(ctx, cfg.AgentModel(), &genai.ClientConfig{
			Project:  cfg.DefaultProjectID(),
			Backend:  genai.BackendVertexAI,
			Location: cfg.DefaultLocation(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Vertex AI model: %w", err)
		}
		return llm, nil

	case "anthropic":
		return anthropic.NewModel(cfg.AnthropicAPIKey(), cfg.AgentModel()), nil

	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", cfg.AgentProvider())
	}
}
