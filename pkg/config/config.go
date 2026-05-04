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

// Package config loads configuration derived from local gcloud defaults.
package config

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// Config contains runtime configuration derived from the environment.
type Config struct {
	userAgent         string
	defaultProjectID  string
	defaultLocation   string
	agentProvider     string
	agentModel        string
	enableDeleteTools bool
	anthropicAPIKey   string
}

// UserAgent returns the user agent string for outbound API calls.
func (c *Config) UserAgent() string {
	return c.userAgent
}

// DefaultProjectID returns the default GCP project ID, if set.
func (c *Config) DefaultProjectID() string {
	return c.defaultProjectID
}

// DefaultLocation returns the default GCP region or zone, if set.
func (c *Config) DefaultLocation() string {
	return c.defaultLocation
}

// AgentProvider returns the configured LLM provider for the agent.
func (c *Config) AgentProvider() string {
	return c.agentProvider
}

// AgentModel returns the configured LLM model for the agent.
func (c *Config) AgentModel() string {
	return c.agentModel
}

// EnableDeleteTools returns true if destructive delete tools are enabled.
func (c *Config) EnableDeleteTools() bool {
	return c.enableDeleteTools
}

// AnthropicAPIKey returns the configured Anthropic API key.
func (c *Config) AnthropicAPIKey() string {
	return c.anthropicAPIKey
}

// New constructs a Config populated from gcloud and build version.
func New(version string, enableDeleteTools bool) *Config {
	provider := os.Getenv("GKE_MCP_PROVIDER")
	if provider == "" {
		provider = "vertex-ai"
	}
	model := os.Getenv("GKE_MCP_MODEL")
	if model == "" {
		if provider == "anthropic" {
			model = "claude-3-7-sonnet-20250219"
		} else {
			model = "gemini-2.5-pro"
		}
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	return &Config{
		userAgent:         "gke-mcp/" + version,
		defaultProjectID:  getDefaultProjectID(),
		defaultLocation:   getDefaultLocation(),
		agentProvider:     provider,
		agentModel:        model,
		enableDeleteTools: enableDeleteTools,
		anthropicAPIKey:   anthropicKey,
	}
}

// NewTestConfig constructs a mock configuration for testing purposes.
func NewTestConfig(project, location, provider, model string) *Config {
	return &Config{
		userAgent:         "gke-mcp/test",
		defaultProjectID:  project,
		defaultLocation:   location,
		agentProvider:     provider,
		agentModel:        model,
		enableDeleteTools: false,
	}
}

func getDefaultProjectID() string {
	projectID, err := getGcloudConfig("core/project")
	if err != nil {
		log.Printf("Failed to get default project: %v", err)
		return ""
	}
	return projectID
}

func getDefaultLocation() string {
	region, err := getGcloudConfig("compute/region")
	if err == nil {
		return region
	}
	zone, err := getGcloudConfig("compute/zone")
	if err == nil {
		return zone
	}
	return ""
}

func getGcloudConfig(key string) (string, error) {
	// #nosec G204
	out, err := exec.Command("gcloud", "config", "get", key).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
