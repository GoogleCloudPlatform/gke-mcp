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
	"log"
	"os/exec"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Install(s *server.MCPServer, _ *config.Config) {
	giqGenerateManifestTool := mcp.NewTool("giq_generate_manifest",
		mcp.WithDescription("Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("model", mcp.Required(), mcp.Description("The model to use. Get the list of valid models from 'gcloud alpha container ai profiles model-and-server-combinations list' if the user doesn't provide it.")),
		mcp.WithString("model_server", mcp.Required(), mcp.Description("The model server to use. Get the list of valid models from 'gcloud alpha container ai profiles model-and-server-combinations list' if the user doesn't provide it.")),
		mcp.WithString("accelerator", mcp.Required(), mcp.Description("The accelerator to use. Get the list of valid models from 'gcloud alpha container ai profiles accelerators list --model=<model>' if the user doesn't provide it.")),
		mcp.WithString("target_ntpot_milliseconds", mcp.Description("The maximum normalized time per output token (NTPOT) in milliseconds.NTPOT is measured as the request_latency / output_tokens.")),
	)
	s.AddTool(giqGenerateManifestTool, giqGenerateManifest)
}

func giqGenerateManifest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	model, err := request.RequireString("model")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	modelServer, err := request.RequireString("model_server")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	accelerator, err := request.RequireString("accelerator")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	targetNtpotMilliseconds := request.GetString("target_ntpot_milliseconds", "")
	args := []string{
		"alpha",
		"container",
		"ai",
		"profiles",
		"manifests",
		"create",
		"--model=" + model,
		"--model-server=" + modelServer,
		"--accelerator-type=" + accelerator,
	}
	if targetNtpotMilliseconds != "" {
		args = append(args, "--target-ntpot-milliseconds="+targetNtpotMilliseconds)
	}
	out, err := exec.Command("gcloud", args...).Output()
	if err != nil {
		log.Printf("Failed to generate manifest: %v", err)
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}
