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
	// Restricts skill and case names to safe alphanumeric characters, hyphens, and underscores to prevent path traversal
	safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

type queryMockRule struct {
	QueryContains string `json:"query_contains,omitempty"`
	Response      string `json:"response,omitempty"`
}

type prometheusMockRule struct {
	QueryContains string          `json:"query_contains,omitempty"`
	Response      json.RawMessage `json:"response,omitempty"`
}

type k8sResourceMockRule struct {
	ResourceType string `json:"resource_type,omitempty"`
	Name         string `json:"name,omitempty"`
	Response     string `json:"response,omitempty"`
}

type caseMockData struct {
	QueryLogs                  []queryMockRule       `json:"query_logs,omitempty"`
	MonitoringTimeSeriesCharts []queryMockRule       `json:"monitoring_time_series_chart,omitempty"`
	Prometheus                 []prometheusMockRule  `json:"prometheus,omitempty"`
	K8sResources               []k8sResourceMockRule `json:"k8s_resources,omitempty"`
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
			res, _, err := handleMockToolCall(ctx, tool.Name, args, c)
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

// handleMockToolCall routes mock tool execution to the appropriate mock data handler.
//
// It resolves the mock scenario identifier (skill name and case name) from the
// configuration (loaded via build-time linker flags or runtime environment variables),
// reads the corresponding case-wide JSON mock data file, and dispatches the call.
//
// If the scenario coordinates cannot be resolved or the mock data file is missing,
// it returns a structured failure indicating missing mock details.
func handleMockToolCall(_ context.Context, toolName string, args any, c *config.Config) (*mcp.CallToolResult, any, error) {
	var envSkill, envCaseName, mockDir string

	if c != nil {
		envSkill = c.MockSkill()
		envCaseName = c.MockCase()
		mockDir = c.MockDataDir()
	} else {
		envSkill = os.Getenv("GKE_MCP_MOCK_SKILL")
		envCaseName = os.Getenv("GKE_MCP_MOCK_CASE")
		if val := os.Getenv("GKE_MCP_MOCK_DATA_DIR"); val != "" {
			mockDir = val
		} else {
			// "mock_data" is the default directory to read from if no override is provided
			mockDir = "mock_data"
		}
	}

	var skill, caseName string
	resolved := false

	if envSkill != "" && envCaseName != "" {
		if safeNameRegex.MatchString(envSkill) && safeNameRegex.MatchString(envCaseName) {
			skill = envSkill
			caseName = envCaseName
			resolved = true
		}
	}

	if !resolved {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("could not resolve mock scenario (skill/case) from environment or tool arguments for tool %s", toolName)}},
		}, nil, nil
	}

	// Load mock_data/<skill>/<caseName>.json
	mockPath := filepath.Join(mockDir, skill, caseName+".json")
	mockBytes, err := os.ReadFile(mockPath) // #nosec G304,G703
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("no mock data file found for skill %q and case %q", skill, caseName)}},
		}, nil, nil
	}

	var data caseMockData
	if err := json.Unmarshal(mockBytes, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	switch toolName {
	case "query_logs":
		query, err := extractQueryArg(args)
		if err != nil {
			return nil, nil, err
		}
		return matchQueryRules(data.QueryLogs, query), nil, nil

	case "monitoring_time_series_chart":
		query, err := extractQueryArg(args)
		if err != nil {
			return nil, nil, err
		}
		return matchQueryRules(data.MonitoringTimeSeriesCharts, query), nil, nil

	case "mql_validator":
		query, err := extractQueryArg(args)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: query},
			},
		}, nil, nil

	case "query_prometheus":
		query, err := extractQueryArg(args)
		if err != nil {
			return nil, nil, err
		}
		res, err := resolveQueryPrometheusMock(mockBytes, query)
		return res, nil, err

	case "get_k8s_resource":
		argsMap, err := extractArgsMap(args)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse arguments: %w", err)
		}
		resourceType, _ := argsMap["resourceType"].(string)
		name, _ := argsMap["name"].(string)
		res, err := resolveK8sResourceMock(mockBytes, resourceType, name)
		return res, nil, err

	case "get_kubeconfig":
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "mock kubeconfig successfully retrieved for target cluster"},
			},
		}, nil, nil

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("no mock implementation available for tool %s", toolName)}},
		}, nil, nil
	}
}

// extractQueryArg extracts the "query" string parameter from generic tool arguments.
func extractQueryArg(args any) (string, error) {
	argsMap, err := extractArgsMap(args)
	if err != nil {
		return "", fmt.Errorf("failed to parse arguments: %w", err)
	}
	query, _ := argsMap["query"].(string)
	return query, nil
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

// matchQueryRules evaluates mock rules sequentially against the query string.
func matchQueryRules(rules []queryMockRule, query string) *mcp.CallToolResult {
	for _, rule := range rules {
		if rule.QueryContains != "" && strings.Contains(query, rule.QueryContains) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: rule.Response},
				},
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("no mock rule matched for query: %s", query)},
		},
	}
}

// resolveQueryPrometheusMock handles simulated output for the query_prometheus tool.
//
// It parses the case-wide JSON data and evaluates prometheus rules sequentially against the PromQL query string.
func resolveQueryPrometheusMock(mockDataBytes []byte, query string) (*mcp.CallToolResult, error) {
	var data caseMockData
	if err := json.Unmarshal(mockDataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	for _, rule := range data.Prometheus {
		if rule.QueryContains != "" && strings.Contains(query, rule.QueryContains) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(rule.Response)},
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

// resolveK8sResourceMock handles simulated output for the get_k8s_resource tool.
//
// It evaluates K8s resource rules sequentially against the resourceType and name arguments,
// utilizing bidirectional expansion matching (+ "s", + "es") for robust singular/plural handling.
func resolveK8sResourceMock(mockDataBytes []byte, resourceType, name string) (*mcp.CallToolResult, error) {
	var data caseMockData
	if err := json.Unmarshal(mockDataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mock data: %w", err)
	}

	for _, rule := range data.K8sResources {
		resourceMatches := rule.ResourceType == "" ||
			strings.EqualFold(rule.ResourceType, resourceType) ||
			strings.EqualFold(rule.ResourceType+"s", resourceType) ||
			strings.EqualFold(rule.ResourceType, resourceType+"s") ||
			strings.EqualFold(rule.ResourceType+"es", resourceType) ||
			strings.EqualFold(rule.ResourceType, resourceType+"es")

		nameMatches := rule.Name == "" || strings.EqualFold(rule.Name, name)

		if resourceMatches && nameMatches {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: rule.Response},
				},
			}, nil
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("no mock rule matched for resource_type: %s, name: %s", resourceType, name)},
		},
	}, nil
}
