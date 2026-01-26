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

// Package k8schangelog provides tools for fetching Kubernetes changelogs.
package k8schangelog

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	kubernetesMinorVersionRegexp = regexp.MustCompile(`^\d+\.\d+$`)
	changelogHostURL             = "https://raw.githubusercontent.com"
)

type getK8sChangelogArgs struct {
	KubernetesMinorVersion string `json:"KubernetesMinorVersion" jsonschema:"The kubernetes minor version to get changelog for. For example, '1.33'."`
}

// Install registers Kubernetes changelog tools with the MCP server.
func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_changelog",
		Description: "Get changelog file for a specific kubernetes minor version and keep only changes content. Prefer to use this tool if kubernetes minor version changelog is needed.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, getK8sChangelog)

	return nil
}

func getK8sChangelog(_ context.Context, _ *mcp.CallToolRequest, args *getK8sChangelogArgs) (*mcp.CallToolResult, any, error) {
	version := strings.TrimSpace(args.KubernetesMinorVersion)
	if !kubernetesMinorVersionRegexp.MatchString(version) {
		return nil, nil, fmt.Errorf("invalid kubernetes minor version: %s", version)
	}

	changelogURL := fmt.Sprintf("%s/kubernetes/kubernetes/refs/heads/master/CHANGELOG/CHANGELOG-%s.md", changelogHostURL, version)
	// #nosec G107
	resp, err := http.Get(changelogURL)
	if err != nil {
		log.Printf("Failed to get changelog: %v", err)
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to get changelog with status code: %d", resp.StatusCode)
		log.Printf("Failed to get changelog: %v", err)
		return nil, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read changelog response body: %v", err)
		return nil, nil, err
	}
	changelogFileContent := string(body)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: keepOnlyChanges(changelogFileContent)},
		},
	}, nil, nil
}

var (
	changelogVersionLineRegexp = regexp.MustCompile(`^# v\d\.\d+\.\d+`)
	ignoredSectionPrefixes     = []string{"## Dependencies", "## Downloads for"}
)

func keepOnlyChanges(changelog string) string {
	var result strings.Builder
	hasMetTheFirstVersionHeading := false // it is set to true only once when the first version heading is met and then never change
	isInIgnoredSection := false
	lines := strings.Split(changelog, "\n")

	for _, line := range lines {
		if !hasMetTheFirstVersionHeading {
			if changelogVersionLineRegexp.MatchString(line) {
				hasMetTheFirstVersionHeading = true
			} else {
				continue
			}
		}

		isIgnoredSectionHeader := false
		for _, prefix := range ignoredSectionPrefixes {
			if strings.HasPrefix(line, prefix) {
				isInIgnoredSection = true
				isIgnoredSectionHeader = true
				break
			}
		}
		if isIgnoredSectionHeader {
			continue
		}

		if isInIgnoredSection {
			if strings.HasPrefix(line, "# ") || strings.HasPrefix(line, "## ") {
				isInIgnoredSection = false
			}
		}

		if !isInIgnoredSection {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	return result.String()
}
