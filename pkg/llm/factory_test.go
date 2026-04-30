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

package llm

import (
	"context"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(context.Background(), nil)

	if err == nil {
		t.Errorf("Expected error for nil config, got nil")
	}
}

func TestNewClient_UnsupportedProvider(t *testing.T) {
	cfg := config.NewTestConfig("test-project", "us-central1", "unsupported-provider", "gemini-2.5-pro")
	_, err := NewClient(context.Background(), cfg)

	if err == nil {
		t.Errorf("Expected error for unsupported provider, got nil")
	}
}

func TestParseProvider_CaseInsensitive(t *testing.T) {
	provider, err := parseProvider("  VeRtEx-AI ")

	if err != nil {
		t.Errorf("Expected no error for mixed-case vertex provider, got %v", err)
	}
	if provider != "vertex-ai" {
		t.Errorf("Expected vertex-ai, got %s", provider)
	}
}
