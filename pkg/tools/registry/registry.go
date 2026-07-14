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

// Package registry handles central MCP tool registration and mock routing.
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	// Extracts cluster_name="..." / cluster_name='...' or cluster "..." / cluster '...' from LQL query
	clusterInQueryRegex = regexp.MustCompile(`(?:cluster_name\s*=\s*|cluster\s+)["']([^"']+)["']`)
	// Restricts skill and case names to safe alphanumeric characters, hyphens, and underscores to prevent path traversal
	safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

type queryLogMockRule struct {
	QueryContains string `json:"query_contains,omitempty"`
	Response      string `json:"response,omitempty"`
}

type caseMockData struct {
	QueryLogs []queryLogMockRule `json:"query_logs,omitempty"`
}

// RegisterTool wraps mcp.AddTool to intercept and mock tool execution in MockMode.
//
// When Config.MockMode() is active, the real tool handler is bypassed, and
// execution is routed through handleClusterEncodedMock to fetch simulated
// telemetry responses from the filesystem workspace. If MockMode is disabled,
// it delegates directly to the production handler.
func RegisterTool[In, Out any](
	s *mcp.Server,
	c *config.Config,
	tool *mcp.Tool,
	handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error),
) {
	mcp.AddTool(s, tool, func(ctx context.Context, req *mcp.CallToolRequest, args In) (*mcp.CallToolResult, Out, error) {
		if c != nil && c.MockMode() {
			res, _, err := handleClusterEncodedMock(ctx, tool.Name, args, c)
			if err != nil {
				var zero Out
				return nil, zero, err
			}
			var zero Out
			return res, zero, nil
		}
		return handler(ctx, req, args)
	})
}

// handleClusterEncodedMock dispatches mock tool calls to their respective mock resolvers.
//
// It dynamically extracts the scenario identifier (skill name and case name)
// from the cluster name passed in parameters or LQL log queries, reads the corresponding
// case JSON mock data, and executes the tool's simulated handler.
//
// If the scenario cannot be resolved or the mock file is missing, it returns
// "no mock data available for this tool under this test case".
func handleClusterEncodedMock(ctx context.Context, toolName string, args any, c *config.Config) (*mcp.CallToolResult, any, error) {
	argsMap, err := extractArgsMap(args)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	var clusterName string
	if val, ok := argsMap["cluster_name"].(string); ok && val != "" {
		clusterName = val
	} else if query, ok := argsMap["query"].(string); ok && query != "" {
		clusterName = extractClusterNameFromQuery(query)
	}

	var skill, caseName string
	resolved := false

	if clusterName != "" {
		var ok bool
		skill, caseName, ok = resolveScenarioFromCluster(clusterName)
		if ok {
			resolved = true
		}
	}

	if !resolved {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("could not resolve cluster name from arguments for tool %s", toolName)}},
		}, nil, nil
	}

	var mockDir string
	if c != nil {
		mockDir = c.MockDataDir()
	}

	// Load mock_data/<skill>/<caseName>.json
	mockPath := filepath.Join(mockDir, skill, caseName+".json")
	mockBytes, err := os.ReadFile(mockPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("no mock data file found for skill %q and case %q", skill, caseName)}},
		}, nil, nil
	}

	if toolName == "query_logs" {
		query, _ := argsMap["query"].(string)
		res, err := resolveQueryLogsMock(mockBytes, query)
		return res, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("no mock implementation available for tool %s", toolName)}},
	}, nil, nil
}

// extractArgsMap deserializes generic tool arguments into a standard map[string]any.
func extractArgsMap(args any) (map[string]any, error) {
	bytes, err := json.Marshal(args)
	if err != nil {
		return nil, err
	}
	var res map[string]any
	err = json.Unmarshal(bytes, &res)
	return res, err
}

// extractClusterNameFromQuery extracts the cluster_name filter value from an LQL string.
func extractClusterNameFromQuery(query string) string {
	matches := clusterInQueryRegex.FindStringSubmatch(query)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// resolveScenarioFromCluster parses the skill and case name from an encoded cluster name.
//
// The GKE cluster name must conform to the convention: "cluster-<skill>--<case_name>",
// where underscores in the case name are normalized to hyphens.
func resolveScenarioFromCluster(clusterName string) (string, string, bool) {
	if !strings.HasPrefix(clusterName, "cluster-") {
		return "", "", false
	}
	name := strings.TrimPrefix(clusterName, "cluster-")
	skill, caseRaw, found := strings.Cut(name, "--")
	if !found || skill == "" || caseRaw == "" {
		return "", "", false
	}
	caseName := strings.ReplaceAll(caseRaw, "-", "_")

	if !safeNameRegex.MatchString(skill) || !safeNameRegex.MatchString(caseName) {
		return "", "", false
	}

	return skill, caseName, true
}

// resolveQueryLogsMock handles simulated output for the query_logs tool.
//
// It parses the case-wide JSON data and evaluates query log rules sequentially against the LQL query string.
func resolveQueryLogsMock(mockDataBytes []byte, query string) (*mcp.CallToolResult, error) {
	var data caseMockData
	if err := json.Unmarshal(mockDataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	for _, rule := range data.QueryLogs {
		if rule.QueryContains != "" && strings.Contains(query, rule.QueryContains) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: rule.Response},
				},
			}, nil
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("no mock rule matched for query: %s", query)},
		},
	}, nil
}


