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
	"net/url"
	"time"
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
	userAgent  string
}

// NewRealDeveloperKnowledgeClient creates a new real client instance.
func NewRealDeveloperKnowledgeClient(baseURL string, apiKey string, userAgent string) *RealDeveloperKnowledgeClient {
	return &RealDeveloperKnowledgeClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:    apiKey,
		userAgent: userAgent,
	}
}

// doPost executes a POST request to the Developer Knowledge API.
func (c *RealDeveloperKnowledgeClient) doPost(ctx context.Context, path string, reqBody interface{}) (string, error) {
	reqURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-Goog-Api-Key", c.apiKey)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		return "", fmt.Errorf("API request failed with status %s: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// doGet executes a GET request to the Developer Knowledge API.
func (c *RealDeveloperKnowledgeClient) doGet(ctx context.Context, path string, queryParams map[string]string) (string, error) {
	reqURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(reqURL)
	if err != nil {
		return "", err
	}
	q := parsedURL.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	parsedURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
	if err != nil {
		return "", err
	}
	if c.apiKey != "" {
		req.Header.Set("X-Goog-Api-Key", c.apiKey)
	}
	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
		return "", fmt.Errorf("API request failed with status %s: %s", resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// GetDocuments fetches specific documents by their IDs.
func (c *RealDeveloperKnowledgeClient) GetDocuments(ctx context.Context, documentIDs []string) (string, error) {
	if len(documentIDs) == 0 {
		return `{"documents": []}`, nil
	}
	return c.doPost(ctx, "/v1/documents:batchGet", map[string]interface{}{
		"names": documentIDs,
	})
}

// AnswerQuery answers a query based on the knowledge base.
func (c *RealDeveloperKnowledgeClient) AnswerQuery(ctx context.Context, query string) (string, error) {
	return c.doPost(ctx, "/v1/TopLevel:answerQuery", map[string]interface{}{
		"query": query,
	})
}

// SearchDocuments searches for documents related to a query.
func (c *RealDeveloperKnowledgeClient) SearchDocuments(ctx context.Context, query string) (string, error) {
	return c.doGet(ctx, "/v1/documents:searchDocumentChunks", map[string]string{
		"query": query,
	})
}
