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

package cost

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/mark3labs/mcp-go/mcp"
)

const gkeCostPromptTemplate = `
You are a GKE cost and optimization expert. Answer the user's question about GKE costs, optimization, or billing using the comprehensive cost context available in the GKE MCP server.
User Question: {{.user_question}}
Based on the GKE cost context available, provide a detailed and helpful response that includes:
1. **Direct Answer**: Address the specific cost question or optimization request
2. **BigQuery Integration**: Explain how to use BigQuery for cost analysis if relevant
3. **Cost Allocation**: Mention GKE Cost Allocation requirements when applicable
4. **Actionable Steps**: Provide concrete next steps or commands when possible
5. **Resource References**: Point to relevant GCP documentation or console links
Key points to remember:
- GKE costs come from GCP Billing Detailed BigQuery Export
- BigQuery CLI (bq) is preferred over BigQuery Studio when available
- GKE Cost Allocation must be enabled for namespace and workload-level cost data
- Required parameters include BigQuery table path, time frame, project ID, cluster details
- Use the cost analysis queries from the GKE MCP documentation as templates
Always be helpful, specific, and actionable in your response.
`

var gkeCostTmpl = template.Must(template.New("gke-cost").Parse(gkeCostPromptTemplate))

// GkeCostPrompt defines the /gke:cost command
var GkeCostPrompt = mcp.NewPrompt("gke:cost",
	mcp.WithPromptDescription("Answer natural language questions about GKE-related costs by leveraging the bundled cost context instructions within the gke-mcp server."),
	mcp.WithArgument("user_question",
		mcp.ArgumentDescription("The user's natural language question about GKE costs"),
		mcp.RequiredArgument(),
	),
)

// GkeCostHandler is the handler function for the /gke:cost prompt
func GkeCostHandler(_ context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	userQuestion := request.Params.Arguments["user_question"]
	if userQuestion == "" {
		return nil, fmt.Errorf("user_question is required")
	}

	var buf bytes.Buffer
	if err := gkeCostTmpl.Execute(&buf, map[string]string{"user_question": userQuestion}); err != nil {
		return nil, fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return mcp.NewGetPromptResult(
		"GKE Cost Analysis Prompt",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser, // Using RoleUser to pass the whole block as user input
				mcp.NewTextContent(buf.String()),
			),
		},
	), nil
}
