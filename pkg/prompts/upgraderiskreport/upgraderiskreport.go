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

// Package upgraderiskreport provides prompt templates for upgrade risk reports.
package upgraderiskreport

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const gkeUpgradeRiskReportPromptTemplate = `
# GKE Upgrade Risk Report Generation

**1. Input Parameters:**
  - Cluster Name: {{.clusterName}}
  - Cluster Location: {{.clusterLocation}}
  - Target Version: {{.targetVersion}}

**2. Your Role:**
You are a GKE expert. Your task is to generate a comprehensive upgrade risk report for the specified GKE cluster, analyzing the potential risks of upgrading from its current version to the 'Target Version'.

**3. Primary Goal:**
Produce a report outlining potential risks, and actionable recommendations to ensure a safe and smooth GKE upgrade. The report should be based on the changes introduced between the cluster's current control plane version and the 'Target Version'.

**4. Handling Missing Target Version:**
If 'Target Version' is not provided:
  a. State that the target version is required.
  b. Use ` + "`gcloud container get-server-config`" + ` to fetch available GKE versions.
  c. Filter this list to show only versions NEWER than the cluster's current control plane version and compatible with the cluster's release channel.
  d. Present these versions to the user to help them choose a 'Target Version'.

**5. Information Gathering & Tools:**
Assume you have the ability to run the following commands to gather necessary information:
  - **Cluster Details:** Use ` + "`gcloud`" + ` to get cluster details like control plane version, release channel, node pool versions, etc.
  - **In-Cluster Resources:** Use ` + "`kubectl`" + ` (after ` + "`gcloud container clusters get-credentials`" + `) for inspecting workloads, APIs in use, etc.
  - **Kubernetes Changelogs:** Use the ` + "`get_k8s_changelog`" + ` tool to fetch kubernetes changelogs.
  - **GKE Release Notes:** Use the ` + "`get_gke_release_notes`" + ` tool to fetch GKE release notes.

**6. Changelog Analysis:**
  - **Minor Versions:** Include changelogs for ALL minor versions from the current control plane minor version up to AND INCLUDING the target minor version. (e.g., 1.29.x to 1.31.y requires looking at changes in 1.29, 1.30, 1.31).
  - **Patch Versions:** Analyze changes for EVERY patch version BETWEEN the current version (exclusive) and the target version (inclusive). (e.g., 1.29.1 to 1.29.5 means analyzing 1.29.2, 1.29.3, 1.29.4, 1.29.5).
  - **GKE Versions:** Analyze changes for GKE version BETWEEN the current version (exclusive) and the target version (inclusive). (e.g., 1.29.1-gke.123000 to 1.29.5-gke.234000 means analyzing 1.29.1-gke.123500, 1.29.1-gke.124000 etc, and 1.29.5-gke.234000).

**7. Risk Identification - Focus on:**
  - **API Deprecations/Removals:** Especially those affecting in-use cluster resources.
  - **Breaking Changes:** Significant behavioral changes in existing, stable features.
  - **Default Configuration Changes:** Modifications to defaults that could alter workload behavior.
  - **New Feature Interactions:** Potentially disruptive interactions between new features and existing setups.
  - Changes REQUIRING manual action before upgrade to prevent outages.

**8. Report Format:**
Present the risks as a single list, ordered by severity. Each risk item MUST follow this markdown structure:

` + "```markdown" + `
# Short Risk Title

## Description

(Detailed description of the change and the potential risk it introduces for THIS specific upgrade)

## Verification Recommendations

(Clear, actionable steps or commands to check if the cluster is affected by this risk. Include example ` + "`kubectl`" + ` or ` + "`gcloud`" + ` commands where appropriate. Reference specific documentation links if possible.)

## Mitigation Recommendations

(Clear, actionable steps, configuration changes, or code adjustments to mitigate the risk BEFORE the upgrade. Provide examples and link to docs.)
` + "```" + `

**9. Principles:**
  - Be specific for each risk; avoid grouping unrelated issues.
  - Ensure Verification and Mitigation steps are practical and provide sufficient detail for a GKE administrator to act upon.
  - Base the analysis SOLELY on the changes between the cluster's current version and the target version.
  - Do not read or write any local files generating the report.
  - In the final report, keep only risks which have mitigation actions, ignore those which have no mitigation actions.

`

var gkeUpgradeRiskReportTmpl = template.Must(template.New("gke-upgrade-risk-report").Parse(gkeUpgradeRiskReportPromptTemplate))

const (
	clusterNameArgName     = "cluster_name"
	clusterLocationArgName = "cluster_location"
	targetVersionArgName   = "target_version"
)

// Install registers the upgrade risk report prompt with the MCP server.
func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	s.AddPrompt(&mcp.Prompt{
		Name:        "gke:upgrade-risk-report",
		Description: "Generate GKE cluster upgrade risk report.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        clusterNameArgName,
				Description: "A name of a GKE cluster user want to upgrade.",
				Required:    true,
			},
			{
				Name:        clusterLocationArgName,
				Description: "A location of a GKE cluster user want to upgrade.",
				Required:    true,
			},
			{
				Name:        targetVersionArgName,
				Description: "A version user want to upgrade their cluster to.",
				Required:    false,
			},
		},
	}, gkeUpgradeRiskReportHandler)

	return nil
}

// gkeUpgradeRiskReportHandler is the handler function for the /gke:upgrade-risk-report prompt
func gkeUpgradeRiskReportHandler(_ context.Context, request *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	clusterName := strings.TrimSpace(request.Params.Arguments[clusterNameArgName])
	if clusterName == "" {
		return nil, fmt.Errorf("argument '%s' cannot be empty", clusterNameArgName)
	}
	clusterLocation := strings.TrimSpace(request.Params.Arguments[clusterLocationArgName])
	if clusterLocation == "" {
		return nil, fmt.Errorf("argument '%s' cannot be empty", clusterLocationArgName)
	}
	targetVersion := strings.TrimSpace(request.Params.Arguments[targetVersionArgName])

	var buf bytes.Buffer
	if err := gkeUpgradeRiskReportTmpl.Execute(&buf, map[string]string{
		"clusterName":     clusterName,
		"clusterLocation": clusterLocation,
		"targetVersion":   targetVersion,
	}); err != nil {
		return nil, fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return &mcp.GetPromptResult{
		Description: "GKE Cluster Upgrade Risk Report Prompt",
		Messages: []*mcp.PromptMessage{
			{
				Content: &mcp.TextContent{
					Text: buf.String(),
				},
				Role: "user",
			},
		},
	}, nil
}
