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

package giq

import (
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type giqGenerateManifestArgs struct {
	Model                   string `json:"model" jsonschema:"The model to use. Get the list of valid models from 'gcloud container ai profiles models list' if the user doesn't provide it."`
	ModelServer             string `json:"model_server" jsonschema:"The model server to use. Get the list of valid model servers from 'gcloud container ai profiles list --format='table(modelServerInfo.model,modelServerInfo.modelServer,modelServerInfo.modelServerVersion,acceleratorType)' if the user doesn't provide it. You can filter that gcloud command on '--model={model}' if the user provides the model."`
	Accelerator             string `json:"accelerator" jsonschema:"The accelerator to use. Get the list of valid accelerators from 'gcloud container ai profiles list --format='table(modelServerInfo.model,modelServerInfo.modelServer,modelServerInfo.modelServerVersion,acceleratorType)' if the user doesn't provide it. You can filter that gcloud command on '--model={model}' and '--model-server={model-server}' if the user provides those values."`
	TargetNTPOTMilliseconds string `json:"target_ntpot_milliseconds,omitempty" jsonschema:"The maximum normalized time per output token (NTPOT) in milliseconds.NTPOT is measured as the request_latency / output_tokens."`
}

func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "giq_generate_manifest",
		Description: "Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:   true,
			IdempotentHint: true,
		},
	}, giqGenerateManifest)

	return nil
}

func giqGenerateManifest(ctx context.Context, req *mcp.CallToolRequest, args *giqGenerateManifestArgs) (*mcp.CallToolResult, any, error) {
	if args.Model == "" {
		return nil, nil, fmt.Errorf("model argument cannot be empty")
	}
	if args.ModelServer == "" {
		return nil, nil, fmt.Errorf("model_server argument cannot be empty")
	}
	if args.Accelerator == "" {
		return nil, nil, fmt.Errorf("accelerator argument cannot be empty")
	}

	gcloudArgs := []string{
		"container",
		"ai",
		"profiles",
		"manifests",
		"create",
		"--model", args.Model,
		"--model-server", args.ModelServer,
		"--accelerator-type", args.Accelerator,
	}
	if args.TargetNTPOTMilliseconds != "" {
		gcloudArgs = append(gcloudArgs, "--target-ntpot-milliseconds", args.TargetNTPOTMilliseconds)
	}
	// #nosec G204
	out, err := exec.Command("gcloud", gcloudArgs...).Output()
	if err != nil {
		log.Printf("Failed to generate manifest: %v", err)

		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(out)},
		},
	}, nil, nil
}
