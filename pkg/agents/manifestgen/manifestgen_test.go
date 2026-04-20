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

package manifestgen

import (
	"context"
	"iter"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

type mockGenerativeModel struct {
	res string
	err error
}

func (m *mockGenerativeModel) Name() string {
	return "mock-model"
}

func (m *mockGenerativeModel) GenerateContent(_ context.Context, _ *model.LLMRequest, _ bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		if m.err != nil {
			yield(nil, m.err)
			return
		}
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Parts: []*genai.Part{
					{Text: m.res},
				},
			},
		}
		yield(resp, nil)
	}
}

func TestNewAgent_NilModel(t *testing.T) {
	_, err := NewAgent(nil, nil)
	if err == nil {
		t.Errorf("Expected error for nil model, got nil")
	}
}

func TestGenerateManifest_Success(t *testing.T) {
	mockModel := &mockGenerativeModel{res: "apiVersion: apps/v1\nkind: Deployment"}
	agent, err := NewAgent(mockModel, &config.Config{})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	manifest, err := agent.Run(context.Background(), "nginx", "test-session")
	if err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if !strings.Contains(manifest, "Deployment") {
		t.Errorf("Expected simulated YAML response, got %v", manifest)
	}
}
