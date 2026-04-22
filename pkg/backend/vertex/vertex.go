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

// Package vertex provides shared connections and models for agents.
package vertex

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/vertexai/genai"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

// Client manages a shared Vertex AI resource pool for all downstream agents.
type Client struct {
	underlying *genai.Client
}

var (
	mu       sync.Mutex
	instance *Client
)

// New initialized a singleton instance of the Google Cloud GenAI client.
func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return instance, nil
	}

	projectID := cfg.DefaultProjectID()
	if projectID == "" {
		return nil, fmt.Errorf("projectID is required in shared connection config")
	}

	location := cfg.DefaultLocation()
	if location == "" {
		location = "us-central1"
	}

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to establish top-level vertex AI pool: %w", err)
	}

	instance = &Client{underlying: client}
	return instance, nil
}

// Model returns a shared or localized instance of a specific text generator.
func (c *Client) Model(modelName string) *genai.GenerativeModel {
	return c.underlying.GenerativeModel(modelName)
}
