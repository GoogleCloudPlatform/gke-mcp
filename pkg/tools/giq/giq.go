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

// Package giq provides tools for GKE Inference Quickstart workflows.
package giq

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GenerateInferenceManifestArgs holds arguments for generating a GKE Inference Quickstart manifest.
type GenerateInferenceManifestArgs struct {
	Model                   string `json:"model" jsonschema:"The model to use. Get the list of valid models from 'gcloud container ai profiles models list' if the user doesn't provide it."`
	ModelServer             string `json:"model_server" jsonschema:"The model server to use. Get the list of valid model servers from 'gcloud container ai profiles list --format='table(modelServerInfo.model,modelServerInfo.modelServer,modelServerInfo.modelServerVersion,acceleratorType)' if the user doesn't provide it. You can filter that gcloud command on '--model={model}' if the user provides the model."`
	Accelerator             string `json:"accelerator" jsonschema:"The accelerator to use. Get the list of valid accelerators from 'gcloud container ai profiles list --format='table(modelServerInfo.model,modelServerInfo.modelServer,modelServerInfo.modelServerVersion,acceleratorType)' if the user doesn't provide it. You can filter that gcloud command on '--model={model}' and '--model-server={model-server}' if the user provides those values."`
	TargetNTPOTMilliseconds string `json:"target_ntpot_milliseconds,omitempty" jsonschema:"The maximum normalized time per output token (NTPOT) in milliseconds.NTPOT is measured as the request_latency / output_tokens."`
}

// Install registers GIQ tools with the MCP server.
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

// GenerateInferenceManifest core logic for GKE Inference Quickstart manifest generation.
func GenerateInferenceManifest(ctx context.Context, args *GenerateInferenceManifestArgs) (string, error) {
	if args == nil {
		return "", fmt.Errorf("args cannot be nil")
	}
	if args.Model == "" {
		return "", fmt.Errorf("model argument cannot be empty")
	}
	if args.ModelServer == "" {
		return "", fmt.Errorf("model_server argument cannot be empty")
	}
	if args.Accelerator == "" {
		return "", fmt.Errorf("accelerator argument cannot be empty")
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
	out, err := exec.CommandContext(ctx, "gcloud", gcloudArgs...).Output()
	if err != nil {
		return "", fmt.Errorf("failed to generate manifest: %w", err)
	}
	return string(out), nil
}

func giqGenerateManifest(ctx context.Context, _ *mcp.CallToolRequest, args *GenerateInferenceManifestArgs) (*mcp.CallToolResult, any, error) {
	manifest, err := GenerateInferenceManifest(ctx, args)
	if err != nil {
		return nil, nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: manifest},
		},
	}, nil, nil
}
