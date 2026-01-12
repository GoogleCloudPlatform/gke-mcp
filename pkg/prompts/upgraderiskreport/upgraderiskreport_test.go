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

package upgraderiskreport

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestGkeUpgradeRiskReportHandler_Success(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
				"target_version":   "1.28.0",
			},
		},
	}

	result, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradeRiskReportHandler() error = %v", err)
	}

	if result == nil {
		t.Fatal("gkeUpgradeRiskReportHandler() returned nil result")
	}

	if len(result.Messages) == 0 {
		t.Error("Expected at least one message in result")
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(content.Text, "my-cluster") {
		t.Errorf("Expected prompt to contain cluster name")
	}
}

func TestGkeUpgradeRiskReportHandler_EmptyClusterName(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "",
				"cluster_location": "us-central1",
				"target_version":   "1.28.0",
			},
		},
	}

	_, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for empty cluster_name, got nil")
	}
}

func TestGkeUpgradeRiskReportHandler_EmptyClusterLocation(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "",
				"target_version":   "1.28.0",
			},
		},
	}

	_, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for empty cluster_location, got nil")
	}
}

func TestGkeUpgradeRiskReportHandler_WhitespaceOnlyClusterName(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "   ",
				"cluster_location": "us-central1",
				"target_version":   "1.28.0",
			},
		},
	}

	_, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for whitespace-only cluster_name, got nil")
	}
}

func TestGkeUpgradeRiskReportHandler_MissingArguments(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{},
		},
	}

	_, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err == nil {
		t.Error("Expected error for missing arguments, got nil")
	}
}

func TestGkeUpgradeRiskReportHandler_PromptDescription(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
				"target_version":   "1.28.0",
			},
		},
	}

	result, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradeRiskReportHandler() error = %v", err)
	}

	if result.Description != "GKE Cluster Upgrade Risk Report Prompt" {
		t.Errorf("Expected description 'GKE Cluster Upgrade Risk Report Prompt', got %s", result.Description)
	}
}

func TestGkeUpgradeRiskReportHandler_MessageRole(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
				"target_version":   "1.28.0",
			},
		},
	}

	result, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradeRiskReportHandler() error = %v", err)
	}

	if result.Messages[0].Role != "user" {
		t.Errorf("Expected message role 'user', got %s", result.Messages[0].Role)
	}
}

func TestGkeUpgradeRiskReportHandler_TemplateSections(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "test-cluster",
				"cluster_location": "us-west1",
				"target_version":   "1.29.0",
			},
		},
	}

	result, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradeRiskReportHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	text := content.Text

	expectedSections := []string{
		"Input Parameters",
		"Risk Identification",
		"Report Format",
	}

	for _, section := range expectedSections {
		if !strings.Contains(text, section) {
			t.Errorf("Expected prompt to contain section %q", section)
		}
	}
}

func TestGkeUpgradeRiskReportHandler_WithoutTargetVersion(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
			},
		},
	}

	result, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradeRiskReportHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if !strings.Contains(content.Text, "my-cluster") {
		t.Error("Expected prompt to contain cluster name")
	}
}

func TestGkeUpgradeRiskReportHandler_NonEmptyResult(t *testing.T) {
	req := &mcp.GetPromptRequest{
		Params: &mcp.GetPromptParams{
			Arguments: map[string]string{
				"cluster_name":     "my-cluster",
				"cluster_location": "us-central1",
				"target_version":   "1.28.0",
			},
		},
	}

	result, err := gkeUpgradeRiskReportHandler(context.Background(), req)
	if err != nil {
		t.Fatalf("gkeUpgradeRiskReportHandler() error = %v", err)
	}

	content := result.Messages[0].Content.(*mcp.TextContent)
	if len(content.Text) == 0 {
		t.Error("Expected non-empty prompt text")
	}
}
