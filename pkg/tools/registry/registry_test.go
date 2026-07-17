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

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

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

func TestHandleMockToolCall_EndToEnd(t *testing.T) {
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

	t.Run("resolve from env variables GKE_MCP_MOCK_SKILL and GKE_MCP_MOCK_CASE", func(t *testing.T) {
		t.Setenv("GKE_MCP_MOCK_SKILL", "my-skill")
		t.Setenv("GKE_MCP_MOCK_CASE", "my_test_case")
		args := map[string]any{
			"query": "error_log",
		}
		res, _, err := handleMockToolCall(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleMockToolCall failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "Found synthetic error log" {
			t.Errorf("Text = %q, want %q", textContent.Text, "Found synthetic error log")
		}
	})

	t.Run("resolve from build-time ldflags (config.BuildMockSkill and config.BuildMockCase)", func(t *testing.T) {
		t.Setenv("GKE_MCP_MOCK_SKILL", "")
		t.Setenv("GKE_MCP_MOCK_CASE", "")
		origSkill := config.BuildMockSkill
		origCase := config.BuildMockCase
		config.BuildMockSkill = "my-skill"
		config.BuildMockCase = "my_test_case"
		defer func() {
			config.BuildMockSkill = origSkill
			config.BuildMockCase = origCase
		}()

		args := map[string]any{
			"query": "error_log",
		}
		res, _, err := handleMockToolCall(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleMockToolCall failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "Found synthetic error log" {
			t.Errorf("Text = %q, want %q", textContent.Text, "Found synthetic error log")
		}
	})

	t.Run("unresolvable mock scenario (both env and args missing)", func(t *testing.T) {
		t.Setenv("GKE_MCP_MOCK_SKILL", "")
		t.Setenv("GKE_MCP_MOCK_CASE", "")
		args := map[string]any{
			"query": `resource.type="k8s_cluster"`,
		}
		res, _, err := handleMockToolCall(ctx, "query_logs", args, cfg)
		if err != nil {
			t.Fatalf("handleMockToolCall failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "could not resolve mock scenario") {
			t.Errorf("Text = %q, want unresolvable scenario error message", textContent.Text)
		}
	})

	t.Run("unsupported tool", func(t *testing.T) {
		t.Setenv("GKE_MCP_MOCK_SKILL", "my-skill")
		t.Setenv("GKE_MCP_MOCK_CASE", "my_test_case")
		args := map[string]any{
			"query": "error_log",
		}
		res, _, err := handleMockToolCall(ctx, "unsupported_tool", args, cfg)
		if err != nil {
			t.Fatalf("handleMockToolCall failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if !strings.Contains(textContent.Text, "no mock implementation available") {
			t.Errorf("Text = %q, want unsupported tool message", textContent.Text)
		}
	})

	t.Run("nil config resolution using env variables fallback", func(t *testing.T) {
		t.Setenv("GKE_MCP_MOCK_SKILL", "my-skill")
		t.Setenv("GKE_MCP_MOCK_CASE", "my_test_case")
		args := map[string]any{
			"query": "error_log",
		}
		res, _, err := handleMockToolCall(ctx, "query_logs", args, nil)
		if err != nil {
			t.Fatalf("handleMockToolCall failed: %v", err)
		}

		textContent := res.Content[0].(*mcp.TextContent)
		if textContent.Text != "Found synthetic error log" {
			t.Errorf("Text = %q, want %q", textContent.Text, "Found synthetic error log")
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
		t.Setenv("GKE_MCP_MOCK_SKILL", "my-skill")
		t.Setenv("GKE_MCP_MOCK_CASE", "my_case")
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
				"cluster_name": "tpu-prod",
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
