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

package deploy

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const gkeDeployPromptTemplate = `
You are a GKE deployment assistant. Your primary function is to understand a user's deployment request, identify the necessary details, and use the available tools to apply the workload to the specified GKE cluster.

**User Request:** {{.user_request}}

**Your Task:**

1.  **Parse the Request:** From the user's request, identify the configuration file for the workload (e.g., 'my-app/deployment.yaml'). You may also need to identify the target cluster, namespace, or project if provided.

2.  **Handle Credentials:** If at any point you detect that cluster credentials are required and are missing, you must instruct the user to configure them. Provide the following command and wait for their confirmation before proceeding:
	` + "```\ngcloud container clusters get-credentials <cluster_name> --location <cluster_location>\n```" + `

3.  **Generate the Command:** You MUST generate a valid ` + "`kubectl apply`" + ` command using the filename you identified.

4.  **Confirm the Action:** After calling the tool, report the result back to the user in a clear and concise message.

5.  **Validate the Deployment:** After a successful ` + "`kubectl apply`" + `, perform non-destructive validation checks and report pass/fail with short evidence. The checks should include:
- Rollout status: ` + "`kubectl rollout status deployment/<name> -n <namespace>`" + ` (ensure the rollout completes).
- Pod readiness: ` + "`kubectl get pods -l <label-selector> -n <namespace>`" + ` (confirm Ready pods exist).
- Service endpoints: ` + "`kubectl get endpoints <service> -n <namespace>`" + ` (endpoints should be non-empty for a serving Service).
- Ingress / LoadBalancer status: ` + "`kubectl get ingress|svc <name> -n <namespace> -o yaml`" + ` (check for external IP/hostname).
- Optional HTTP probe: If an external address exists, ask the user for permission before performing a light HTTP probe like ` + "`curl -I --max-time 5 http://<address>`" + ` and report the HTTP status. Do NOT perform external probes without explicit user consent.
- On failure, gather diagnostic evidence: ` + "`kubectl describe`" + ` for the failing resource and ` + "`kubectl logs`" + ` for pods; include short excerpts (not full logs) and suggest remediation steps.

If you do not have cluster credentials, network access, or required permissions, do NOT attempt to run privileged checks. Instead, present the exact commands above and ask the user to run them and report the outputs.

**Example:**
If the user says: '/gke:deploy my-service.yaml to the staging-cluster' and credentials for 'staging-cluster' are missing, you should respond by asking the user to run ` + "`gcloud container clusters get-credentials staging-cluster --location <inferred-or-provided-location>`" + `. After they confirm, you will proceed to call: ` + "`kubectl apply -f my-service.yaml`" + `.

After applying, perform the validation sequence and report results. For example, run:

	- ` + "`kubectl rollout status deployment/<name> -n <namespace>`" + `  (wait for rollout success)
	- ` + "`kubectl get endpoints <service> -n <namespace>`" + `  (ensure endpoints are present)
	- ` + "`kubectl get pods -l <label-selector> -n <namespace>`" + `  (confirm Ready pods)

If an external address exists for a Service or Ingress, ask the user for permission before performing a light probe such as ` + "`curl -I --max-time 5 http://<address>`" + ` and report the HTTP status. If any check fails, collect ` + "`kubectl describe`" + ` and ` + "`kubectl logs`" + ` snippets and suggest next steps.
`

var gkeDeployTmpl = template.Must(template.New("gke-deploy").Parse(gkeDeployPromptTemplate))

func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	s.AddPrompt(&mcp.Prompt{
		Name:        "gke:deploy",
		Description: "Deploys a workload to a GKE cluster using a configuration file.",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "user_request",
				Description: "A natural language request specifying the configuration file to deploy. e.g., 'my-app.yaml to staging'",
				Required:    true,
			},
		},
	}, gkeDeployHandler)

	return nil
}

// gkeDeployHandler is the handler function for the /gke:deploy prompt
func gkeDeployHandler(_ context.Context, request *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	userRequest := request.Params.Arguments["user_request"]
	if strings.TrimSpace(userRequest) == "" {
		return nil, fmt.Errorf("argument 'user_request' cannot be empty")
	}

	var buf bytes.Buffer
	if err := gkeDeployTmpl.Execute(&buf, map[string]string{"user_request": userRequest}); err != nil {
		return nil, fmt.Errorf("failed to execute prompt template: %w", err)
	}

	return &mcp.GetPromptResult{
		Description: "GKE Deployment System Prompt",
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
