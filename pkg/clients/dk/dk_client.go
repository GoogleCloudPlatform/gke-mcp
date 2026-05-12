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

// Package dk provides the Developer Knowledge API client.
package dk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DeveloperKnowledgeClient defines the interface for interacting with the Developer Knowledge API.
type DeveloperKnowledgeClient interface {
	GetDocuments(ctx context.Context, documentIDs []string) (string, error)
	AnswerQuery(ctx context.Context, query string) (string, error)
	SearchDocuments(ctx context.Context, query string) (string, error)
}

// RealDeveloperKnowledgeClient is the actual implementation.
type RealDeveloperKnowledgeClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewRealDeveloperKnowledgeClient creates a new real client instance.
func NewRealDeveloperKnowledgeClient(baseURL string, apiKey string) *RealDeveloperKnowledgeClient {
	if baseURL == "" {
		baseURL = "https://knowledge.googleapis.com"
	}
	return &RealDeveloperKnowledgeClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		apiKey:     apiKey,
	}
}

// GetDocuments fetches specific documents by their IDs.
func (c *RealDeveloperKnowledgeClient) GetDocuments(_ context.Context, _ []string) (string, error) {
	return "", fmt.Errorf("GetDocuments not implemented")
}

// AnswerQuery answers a query based on the knowledge base.
func (c *RealDeveloperKnowledgeClient) AnswerQuery(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("AnswerQuery not implemented")
}

// SearchDocuments searches for documents related to a query.
func (c *RealDeveloperKnowledgeClient) SearchDocuments(ctx context.Context, query string) (string, error) {
	url := fmt.Sprintf("%s/v1/documents:searchDocumentChunks", c.baseURL)

	reqBody, err := json.Marshal(map[string]interface{}{
		"query": query,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Goog-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %s: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
