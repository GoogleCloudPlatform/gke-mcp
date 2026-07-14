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

package registry

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestExtractClusterNameFromQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "double quotes cluster_name",
			query:    `resource.type="k8s_cluster" AND resource.labels.cluster_name="cluster-my-skill--case-one"`,
			expected: "cluster-my-skill--case-one",
		},
		{
			name:     "single quotes cluster_name",
			query:    `resource.type='k8s_cluster' AND resource.labels.cluster_name='cluster-my-skill--case-one'`,
			expected: "cluster-my-skill--case-one",
		},
		{
			name:     "single quotes cluster",
			query:    `resource.type='k8s_cluster' AND cluster 'cluster-my-skill--case-two'`,
			expected: "cluster-my-skill--case-two",
		},
		{
			name:     "double quotes cluster",
			query:    `resource.type="k8s_cluster" AND cluster "cluster-my-skill--case-two"`,
			expected: "cluster-my-skill--case-two",
		},
		{
			name:     "query without cluster name",
			query:    `resource.type="k8s_container" AND textPayload:"error"`,
			expected: "",
		},
		{
			name:     "empty query",
			query:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractClusterNameFromQuery(tt.query)
			if got != tt.expected {
				t.Errorf("extractClusterNameFromQuery() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestResolveScenarioFromCluster(t *testing.T) {
	tests := []struct {
		name         string
		clusterName  string
		wantSkill    string
		wantCaseName string
		wantOk       bool
	}{
		{
			name:         "valid encoded cluster name",
			clusterName:  "cluster-gke-ai-troubleshooting-tpu-connection-failure-vbar-oom--no-vbar-error-found",
			wantSkill:    "gke-ai-troubleshooting-tpu-connection-failure-vbar-oom",
			wantCaseName: "no_vbar_error_found",
			wantOk:       true,
		},
		{
			name:         "case name with multiple hyphens",
			clusterName:  "cluster-my-skill--oom-without-custom-metrics",
			wantSkill:    "my-skill",
			wantCaseName: "oom_without_custom_metrics",
			wantOk:       true,
		},
		{
			name:         "missing cluster prefix",
			clusterName:  "my-skill--case-name",
			wantSkill:    "",
			wantCaseName: "",
			wantOk:       false,
		},
		{
			name:         "missing double hyphen separator",
			clusterName:  "cluster-my-skill-case-name",
			wantSkill:    "",
			wantCaseName: "",
			wantOk:       false,
		},
		{
			name:         "path traversal rejection",
			clusterName:  "cluster-../../etc/passwd--case-name",
			wantSkill:    "",
			wantCaseName: "",
			wantOk:       false,
		},
		{
			name:         "invalid characters rejection",
			clusterName:  "cluster-skill$name--case*name",
			wantSkill:    "",
			wantCaseName: "",
			wantOk:       false,
		},
		{
			name:         "empty cluster name",
			clusterName:  "",
			wantSkill:    "",
			wantCaseName: "",
			wantOk:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSkill, gotCase, ok := resolveScenarioFromCluster(tt.clusterName)
			if ok != tt.wantOk {
				t.Fatalf("resolveScenarioFromCluster() ok = %v, want %v", ok, tt.wantOk)
			}
			if gotSkill != tt.wantSkill {
				t.Errorf("resolveScenarioFromCluster() skill = %q, want %q", gotSkill, tt.wantSkill)
			}
			if gotCase != tt.wantCaseName {
				t.Errorf("resolveScenarioFromCluster() case = %q, want %q", gotCase, tt.wantCaseName)
			}
		})
	}
}

func TestExtractArgsMap(t *testing.T) {
	type dummyArgs struct {
		ClusterName string `json:"cluster_name"`
		Query       string `json:"query"`
	}

	args := dummyArgs{
		ClusterName: "cluster-foo--bar",
		Query:       "resource.type=k8s_cluster",
	}

	argsMap, err := extractArgsMap(args)
	if err != nil {
		t.Fatalf("extractArgsMap failed: %v", err)
	}

	if got := argsMap["cluster_name"]; got != "cluster-foo--bar" {
		t.Errorf("argsMap[\"cluster_name\"] = %v, want %q", got, "cluster-foo--bar")
	}
	if got := argsMap["query"]; got != "resource.type=k8s_cluster" {
		t.Errorf("argsMap[\"query\"] = %v, want %q", got, "resource.type=k8s_cluster")
	}
}

func TestResolveQueryLogsMock(t *testing.T) {
	mockJSON := `{
		"query_logs": [
			{
				"query_contains": "vbar_control_agent",
				"response": "I0702 OOM detected in vbar_control_agent"
			},
			{
				"query_contains": "device_plugin",
				"response": "Device plugin healthy"
			}
		]
	}`

	t.Run("matching query rule", func(t *testing.T) {
		res, err := resolveQueryLogsMock([]byte(mockJSON), `resource.labels.cluster_name="c" AND "vbar_control_agent"`)
		if err != nil {
			t.Fatalf("resolveQueryLogsMock failed: %v", err)
		}

		if len(res.Content) == 0 {
			t.Fatal("expected content in CallToolResult")
		}

		textContent, ok := res.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("expected *mcp.TextContent, got %T", res.Content[0])
		}

		if textContent.Text != "I0702 OOM detected in vbar_control_agent" {
			t.Errorf("textContent.Text = %q, want expected log response", textContent.Text)
		}
	})

	t.Run("unmatched query rule", func(t *testing.T) {
		res, err := resolveQueryLogsMock([]byte(mockJSON), `resource.type="k8s_node" AND "unknown_log"`)
		if err != nil {
			t.Fatalf("resolveQueryLogsMock failed: %v", err)
		}

		textContent, ok := res.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("expected *mcp.TextContent, got %T", res.Content[0])
		}

		if !strings.Contains(textContent.Text, "no mock rule matched for query") {
			t.Errorf("textContent.Text = %q, want unmatched error string", textContent.Text)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := resolveQueryLogsMock([]byte(`invalid json`), `query`)
		if err == nil {
			t.Error("expected error for invalid json, got nil")
		}
	})
}

func TestHandleClusterEncodedMock_EndToEnd(t *testing.T) {
	// Set up temporary mock_data directory structure
	tempDir := t.TempDir()
	skillDir := filepath.Join(tempDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}

	caseMockContent := `{
		"query_logs": [
			{
				"query_contains": "error_log",
				"response": "Found synthetic error log"
			}
		]
	}`
	mockFilePath := filepath.Join(skillDir, "my_test_case.json")
	if err := os.WriteFile(mockFilePath, []byte(caseMockContent), 0644); err != nil {
		t.Fatalf("failed to write mock file: %v", err)
	}

	t.Setenv("GKE_MCP_MOCK", "true")
	t.Setenv("GKE_MCP_MOCK_DATA_DIR", tempDir)
	cfg := config.New("test", false)

	ctx := context.Background()

	t.Run("resolve from cluster_name argument parameter", func(t *testing.T) {
		args := map[string]any{
			"cluster_name": "cluster-my-skill--my-test-case",
			"query":        "error_log",
		}
		res, _, err := handleClusterEncodedMock(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleClusterEncodedMock failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "Found synthetic error log" {
			t.Errorf("Text = %q, want %q", textContent.Text, "Found synthetic error log")
		}
	})

	t.Run("resolve from query filter", func(t *testing.T) {
		args := map[string]any{
			"query": `resource.labels.cluster_name="cluster-my-skill--my-test-case" AND "error_log"`,
		}
		res, _, err := handleClusterEncodedMock(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleClusterEncodedMock failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "Found synthetic error log" {
			t.Errorf("Text = %q, want %q", textContent.Text, "Found synthetic error log")
		}
	})

	t.Run("unresolvable cluster name", func(t *testing.T) {
		args := map[string]any{
			"query": `resource.type="k8s_cluster"`,
		}
		res, _, err := handleClusterEncodedMock(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleClusterEncodedMock failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "could not resolve cluster name") {
			t.Errorf("Text = %q, want unresolvable cluster error message", textContent.Text)
		}
	})

	t.Run("resolvable cluster but missing mock file", func(t *testing.T) {
		args := map[string]any{
			"cluster_name": "cluster-my-skill--non-existent-case",
		}
		res, _, err := handleClusterEncodedMock(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleClusterEncodedMock failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "no mock data file found") {
			t.Errorf("Text = %q, want missing mock file error message", textContent.Text)
		}
	})

	t.Run("unsupported tool", func(t *testing.T) {
		args := map[string]any{
			"cluster_name": "cluster-my-skill--my-test-case",
		}
		res, _, err := handleClusterEncodedMock(ctx, "unsupported_tool", args, cfg)
		if err != nil {
			t.Fatalf("handleClusterEncodedMock failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "no mock implementation available") {
			t.Errorf("Text = %q, want unsupported tool message", textContent.Text)
		}
	})
}

func TestRegisterTool_MockModeToggle(t *testing.T) {
	tempDir := t.TempDir()
	skillDir := filepath.Join(tempDir, "my-skill")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatalf("failed to create skill dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "my_case.json"), []byte(`{
		"query_logs": [{"query_contains": "error", "response": "mock response"}]
	}`), 0644); err != nil {
		t.Fatalf("failed to write mock file: %v", err)
	}

	tool := &mcp.Tool{
		Name:        "query_logs",
		Description: "test tool",
	}

	type dummyArgs struct {
		ClusterName string `json:"cluster_name"`
		Query       string `json:"query"`
	}

	t.Run("production mode delegates to prod handler", func(t *testing.T) {
		prodHandlerCalled := false
		prodHandler := func(ctx context.Context, req *mcp.CallToolRequest, args dummyArgs) (*mcp.CallToolResult, any, error) {
			prodHandlerCalled = true
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "prod response"}},
			}, nil, nil
		}

		t.Setenv("GKE_MCP_MOCK", "false")
		cfg := config.New("test", false)

		server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
		RegisterTool(server, cfg, tool, prodHandler)

		clientTransport, serverTransport := mcp.NewInMemoryTransports()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			_ = server.Run(ctx, serverTransport)
		}()

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
		session, err := client.Connect(ctx, clientTransport, nil)
		if err != nil {
			t.Fatalf("failed to connect client: %v", err)
		}
		defer session.Close()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "query_logs",
			Arguments: map[string]any{
				"cluster_name": "cluster-my-skill--my-case",
				"query":        "error",
			},
		})
		if err != nil {
			t.Fatalf("CallTool failed: %v", err)
		}

		if !prodHandlerCalled {
			t.Errorf("expected prodHandlerCalled to be true in production mode")
		}

		if len(res.Content) == 0 {
			t.Fatal("expected content in res")
		}
		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "prod response" {
			t.Errorf("got text %q, want %q", textContent.Text, "prod response")
		}
	})

	t.Run("mock mode intercepts call", func(t *testing.T) {
		prodHandlerCalled := false
		prodHandler := func(ctx context.Context, req *mcp.CallToolRequest, args dummyArgs) (*mcp.CallToolResult, any, error) {
			prodHandlerCalled = true
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "prod response"}},
			}, nil, nil
		}

		t.Setenv("GKE_MCP_MOCK", "true")
		t.Setenv("GKE_MCP_MOCK_DATA_DIR", tempDir)
		cfg := config.New("test", false)

		server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
		RegisterTool(server, cfg, tool, prodHandler)

		clientTransport, serverTransport := mcp.NewInMemoryTransports()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			_ = server.Run(ctx, serverTransport)
		}()

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
		session, err := client.Connect(ctx, clientTransport, nil)
		if err != nil {
			t.Fatalf("failed to connect client: %v", err)
		}
		defer session.Close()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "query_logs",
			Arguments: map[string]any{
				"cluster_name": "cluster-my-skill--my-case",
				"query":        "error",
			},
		})
		if err != nil {
			t.Fatalf("CallTool failed: %v", err)
		}

		if prodHandlerCalled {
			t.Errorf("expected prodHandlerCalled to be false in mock mode")
		}

		if len(res.Content) == 0 {
			t.Fatal("expected content in res")
		}
		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "mock response" {
			t.Errorf("got text %q, want %q", textContent.Text, "mock response")
		}
	})

	t.Run("nil config safely delegates to prod handler", func(t *testing.T) {
		prodHandlerCalled := false
		prodHandler := func(ctx context.Context, req *mcp.CallToolRequest, args dummyArgs) (*mcp.CallToolResult, any, error) {
			prodHandlerCalled = true
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "prod response"}},
			}, nil, nil
		}

		server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
		RegisterTool(server, nil, tool, prodHandler)

		clientTransport, serverTransport := mcp.NewInMemoryTransports()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			_ = server.Run(ctx, serverTransport)
		}()

		client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
		session, err := client.Connect(ctx, clientTransport, nil)
		if err != nil {
			t.Fatalf("failed to connect client: %v", err)
		}
		defer session.Close()

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "query_logs",
			Arguments: map[string]any{
				"cluster_name": "cluster-my-skill--my-case",
				"query":        "error",
			},
		})
		if err != nil {
			t.Fatalf("CallTool failed: %v", err)
		}

		if !prodHandlerCalled {
			t.Errorf("expected prodHandlerCalled to be true when config is nil")
		}

		if len(res.Content) == 0 {
			t.Fatal("expected content in res")
		}
		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "prod response" {
			t.Errorf("got text %q, want %q", textContent.Text, "prod response")
		}
	})
}
