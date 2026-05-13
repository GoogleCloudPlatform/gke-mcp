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

package dk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// MockDeveloperKnowledgeClient is a mock implementation for testing.
type MockDeveloperKnowledgeClient struct{}

func (m *MockDeveloperKnowledgeClient) GetDocuments(_ context.Context, documentIDs []string) (string, error) {
	return fmt.Sprintf("Mock documents for IDs: %v", documentIDs), nil
}

func (m *MockDeveloperKnowledgeClient) AnswerQuery(_ context.Context, query string) (string, error) {
	return fmt.Sprintf("Mock answer for query: %s", query), nil
}

func (m *MockDeveloperKnowledgeClient) SearchDocuments(_ context.Context, query string) (string, error) {
	return fmt.Sprintf("Mock search results for query: %s", query), nil
}

func TestRealDeveloperKnowledgeClient_SearchDocuments(t *testing.T) {
	expectedQuery := "gke network policy"
	mockResponse := `{"results": [{"chunk": "details"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/documents:searchDocumentChunks" {
			t.Errorf("Expected path /v1/documents:searchDocumentChunks, got %s", r.URL.Path)
		}
		if r.Header.Get("X-Goog-Api-Key") != "test-api-key" {
			t.Errorf("Expected API Key header, got %s", r.Header.Get("X-Goog-Api-Key"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("User-Agent") != "gke-mcp/test" {
			t.Errorf("Expected User-Agent gke-mcp/test, got %s", r.Header.Get("User-Agent"))
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if body["query"] != expectedQuery {
			t.Errorf("Expected query %q, got %q", expectedQuery, body["query"])
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	client := NewRealDeveloperKnowledgeClient(server.URL, "test-api-key", "gke-mcp/test")
	resp, err := client.SearchDocuments(context.Background(), expectedQuery)
	if err != nil {
		t.Fatalf("SearchDocuments failed: %v", err)
	}
	if resp != mockResponse {
		t.Errorf("Expected response %s, got %s", mockResponse, resp)
	}
}

func TestRealDeveloperKnowledgeClient_SearchDocuments_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewRealDeveloperKnowledgeClient(server.URL, "test-api-key", "gke-mcp/test")
	_, err := client.SearchDocuments(context.Background(), "gke network policy")
	if err == nil {
		t.Fatalf("Expected error for non-200 status code, got nil")
	}
	expectedErrSubstring := "API request failed with status 500 Internal Server Error: Internal Server Error"
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Errorf("Expected error containing %q, got %v", expectedErrSubstring, err)
	}
}
