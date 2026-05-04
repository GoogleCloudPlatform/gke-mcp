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

package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

func TestGenerateContent_Success(t *testing.T) {
	// Mock Anthropic API response
	mockResp := anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{
				Type: "text",
				Text: "Generated manifest here",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.URL.Path != "/v1/messages" {
			t.Errorf("Expected path /v1/messages, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		// Read body to verify it
		body, _ := io.ReadAll(r.Body)
		var params anthropic.MessageNewParams
		if err := json.Unmarshal(body, &params); err != nil {
			t.Errorf("Failed to unmarshal request body: %v", err)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	}))
	defer server.Close()

	// Create model with mock server
	m := &AnthropicModel{
		client: anthropic.NewClient(
			option.WithAPIKey("dummy"),
			option.WithHTTPClient(server.Client()),
			option.WithBaseURL(server.URL),
		),
		model: "claude-3-7-sonnet-20250219",
	}

	req := &model.LLMRequest{
		Contents: []*genai.Content{
			{
				Role: "user",
				Parts: []*genai.Part{
					{
						Text: "Create a deployment",
					},
				},
			},
		},
	}

	seq := m.GenerateContent(context.Background(), req, false)

	var found bool
	for resp, err := range seq {
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		found = true
		if resp.Content.Parts[0].Text != "Generated manifest here" {
			t.Errorf("Expected 'Generated manifest here', got %s", resp.Content.Parts[0].Text)
		}
	}

	if !found {
		t.Errorf("Expected at least one response from iterator")
	}
}
