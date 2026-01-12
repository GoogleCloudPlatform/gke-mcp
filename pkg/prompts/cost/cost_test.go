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

package cost

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGkeCostHandler_Success(t *testing.T) {
	userQuestion := "How do I optimize my GKE costs?"
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": userQuestion,
			},
		},
	}

	result, err := gkeCostHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeCostHandler() error = %v", err)
	}

	if result == nil {
		t.Fatal("gkeCostHandler() returned nil result")
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

	if !strings.Contains(textContent.Text, userQuestion) {
		t.Errorf("Expected prompt to contain user question %q", userQuestion)
	}
}

func TestGkeCostHandler_EmptyQuestion(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": "",
			},
		},
	}

	_, err := gkeCostHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for empty user question, got nil")
	}

	expectedErrMsg := "argument 'user_question' cannot be empty"
	if err.Error() != expectedErrMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedErrMsg)
	}
}

func TestGkeCostHandler_WhitespaceOnlyQuestion(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": "   ",
			},
		},
	}

	_, err := gkeCostHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for whitespace-only user question, got nil")
	}
}

func TestGkeCostHandler_MissingArgument(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{},
		},
	}

	_, err := gkeCostHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for missing user_question argument, got nil")
	}
}

func TestGkeCostHandler_PromptDescription(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": "test question",
			},
		},
	}

	result, err := gkeCostHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeCostHandler() error = %v", err)
	}

	if result.Description != "GKE Cost Analysis Prompt" {
		t.Errorf("Expected description 'GKE Cost Analysis Prompt', got %s", result.Description)
	}
}

func TestGkeCostHandler_MessageRole(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": "test question",
			},
		},
	}

	result, err := gkeCostHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeCostHandler() error = %v", err)
	}

	if len(result.Messages) == 0 {
		t.Fatal("Expected at least one message")
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("Expected message role 'user', got %s", result.Messages[0].Role)
	}
}

func TestGkeCostHandler_TemplateContainsExpectedSections(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": "What are the cost optimization strategies?",
			},
		},
	}

	result, err := gkeCostHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeCostHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	expectedSections := []string{
		"BigQuery Integration",
		"Cost Allocation",
		"Actionable Steps",
	}

	for _, section := range expectedSections {
		if !strings.Contains(text, section) {
			t.Errorf("Expected prompt to contain section %q", section)
		}
	}
}

func TestGkeCostHandler_LongQuestion(t *testing.T) {
	longQuestion := strings.Repeat("How do I optimize my GKE costs? ", 100)
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": longQuestion,
			},
		},
	}

	result, err := gkeCostHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeCostHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(content.Text, longQuestion) {
		t.Error("Expected prompt to contain the full long question")
	}
}

func TestGkeCostHandler_NonEmptyResult(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"user_question": "test",
			},
		},
	}

	result, err := gkeCostHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeCostHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if len(content.Text) == 0 {
		t.Error("Expected non-empty prompt text")
	}
}
