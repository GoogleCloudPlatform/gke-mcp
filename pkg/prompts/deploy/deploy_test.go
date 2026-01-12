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

package deploy

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGkeDeployHandler_Success(t *testing.T) {
	userRequest := "Deploy my-app.yaml to staging"
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": userRequest,
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	if result == nil {
		t.Fatal("gkeDeployHandler() returned nil result")
	}

	if len(result.Messages) == 0 {
		t.Error("Expected at least one message in result")
	}

	content := result.Messages[0].Content
	if content == nil {
		t.Fatal("Expected content in first message")
	}

	textContent, ok := content.(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected TextContent, got %T", content)
	}

	if !strings.Contains(textContent.Text, "GKE") {
		t.Errorf("Expected prompt to contain 'GKE'")
	}
}

func TestGkeDeployHandler_EmptyRequest(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "",
			},
		},
	}

	_, err := gkeDeployHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for empty user request, got nil")
	}

	expectedErrMsg := "argument 'user_request' cannot be empty"
	if err.Error() != expectedErrMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedErrMsg)
	}
}

func TestGkeDeployHandler_WhitespaceOnlyRequest(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "   ",
			},
		},
	}

	_, err := gkeDeployHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for whitespace-only user request, got nil")
	}
}

func TestGkeDeployHandler_MissingArgument(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{},
		},
	}

	_, err := gkeDeployHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for missing user_request argument, got nil")
	}
}

func TestGkeDeployHandler_PromptDescription(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "test request",
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	if result.Description != "GKE Deployment System Prompt" {
		t.Errorf("Expected description 'GKE Deployment System Prompt', got %s", result.Description)
	}
}

func TestGkeDeployHandler_MessageRole(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "test request",
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	if len(result.Messages) == 0 {
		t.Fatal("Expected at least one message")
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("Expected message role 'user', got %s", result.Messages[0].Role)
	}
}

func TestGkeDeployHandler_WorkflowSections(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "Deploy application",
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	expectedSections := []string{
		"Initial Assessment",
		"Guided Execution",
		"Verification",
	}

	for _, section := range expectedSections {
		if !strings.Contains(text, section) {
			t.Errorf("Expected prompt to contain section %q", section)
		}
	}
}

func TestGkeDeployHandler_Idempotency(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "Deploy application",
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	if !strings.Contains(text, "Idempotency") {
		t.Error("Expected prompt to mention idempotency")
	}
}

func TestGkeDeployHandler_ConversationalInteraction(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "Help me deploy",
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	if !strings.Contains(text, "Natural Language") {
		t.Error("Expected prompt to mention natural language interaction")
	}
}

func TestGkeDeployHandler_NonEmptyResult(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_request": "test",
			},
		},
	}

	result, err := gkeDeployHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeDeployHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if len(content.Text) == 0 {
		t.Error("Expected non-empty prompt text")
	}
}
