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

package upgradesbestpracticesriskreport

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGkeUpgradesBestPracticesRiskReportHandler_Success(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
			},
		},
	}

	result, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradesBestPracticesRiskReportHandler() error = %v", err)
	}

	if result == nil {
		t.Fatal("gkeUpgradesBestPracticesRiskReportHandler() returned nil result")
	}

	if len(result.Messages) == 0 {
		t.Error("Expected at least one message in result")
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(content.Text, "my-cluster") {
		t.Errorf("Expected prompt to contain cluster name")
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_EmptyClusterName(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "",
				"cluster_location": "us-central1",
			},
		},
	}

	_, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for empty cluster_name, got nil")
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_EmptyClusterLocation(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "",
			},
		},
	}

	_, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for empty cluster_location, got nil")
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_WhitespaceOnlyClusterName(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "   ",
				"cluster_location": "us-central1",
			},
		},
	}

	_, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for whitespace-only cluster_name, got nil")
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_MissingArguments(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{},
		},
	}

	_, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_PromptDescription(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
			},
		},
	}

	result, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradesBestPracticesRiskReportHandler() error = %v", err)
	}

	if result.Description != "GKE Cluster Upgrade Best Practices Risk Report Prompt" {
		t.Errorf("Expected description 'GKE Cluster Upgrade Best Practices Risk Report Prompt', got %s", result.Description)
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_MessageRole(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
			},
		},
	}

	result, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradesBestPracticesRiskReportHandler() error = %v", err)
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("Expected message role 'user', got %s", result.Messages[0].Role)
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_TemplateSections(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "test-cluster",
				"cluster_location": "us-west1",
			},
		},
	}

	result, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradesBestPracticesRiskReportHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	expectedSections := []string{
		"Maintenance Windows",
		"Pod Disruption Budgets",
		"Node Pool Upgrades",
	}

	for _, section := range expectedSections {
		if !strings.Contains(text, section) {
			t.Errorf("Expected prompt to contain section %q", section)
		}
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_NonEmptyResult(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
			},
		},
	}

	result, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradesBestPracticesRiskReportHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if len(content.Text) == 0 {
		t.Error("Expected non-empty prompt text")
	}
}

func TestGkeUpgradesBestPracticesRiskReportHandler_MitigationRecommendations(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
			},
		},
	}

	result, err := gkeUpgradesBestPracticesRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradesBestPracticesRiskReportHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	if !strings.Contains(text, "Mitigation Recommendations") {
		t.Error("Expected prompt to contain Mitigation Recommendations section")
	}
}
