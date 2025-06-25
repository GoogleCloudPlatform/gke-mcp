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

package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	defaultProjectID string
)

func main() {

	defaultProjectID = getDefaultProject()

	// Create a new MCP server
	s := server.NewMCPServer(
		"GKE MCP Server",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	listClustersTool := mcp.NewTool("list_clusters",
		mcp.WithDescription("List GKE clusters. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("project_id", mcp.DefaultString(defaultProjectID), mcp.Description("GCP project ID. Use the default if the user doesn't provide it.")),
		mcp.WithString("location", mcp.Description("GKE cluster location. Leave this empty if the user doesn't doesn't provide it.")),
	)
	s.AddTool(listClustersTool, listClusters)

	getClusterTool := mcp.NewTool("get_cluster",
		mcp.WithDescription("Get / describe a GKE cluster. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("project_id", mcp.DefaultString(defaultProjectID), mcp.Description("GCP project ID. Use the default if the user doesn't provide it.")),
		mcp.WithString("location", mcp.Required(), mcp.Description("GKE cluster location. Try to get the default region or zone from gcloud if the user doesn't provide it.")),
		mcp.WithString("name", mcp.Required(), mcp.Description("GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name.")),
	)
	s.AddTool(getClusterTool, getCluster)

	giqGenerateManifestTol := mcp.NewTool("giq_generate_manifest",
		mcp.WithDescription("Use Google Inference Quickstart to generate a Kubernetes manifest for AI / inference workloads. Prefer to use this tool instead of gcloud"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("model", mcp.Required(), mcp.Description("The model to use. Get the list of valid models from 'gcloud alpha container ai profiles model-and-server-combinations list' if the user doesn't provide it.")),
		mcp.WithString("model_server", mcp.Required(), mcp.Description("The model server to use. Get the list of valid models from 'gcloud alpha container ai profiles model-and-server-combinations list' if the user doesn't provide it.")),
		mcp.WithString("accelerator", mcp.Required(), mcp.Description("The accelerator to use. Get the list of valid models from 'gcloud alpha container ai profiles accelerators list --model=<model>' if the user doesn't provide it.")),
		mcp.WithString("target_ntpot_milliseconds", mcp.Description("The maximum normalized time per output token (NTPOT) in milliseconds.NTPOT is measured as the request_latency / output_tokens.")),
	)
	s.AddTool(giqGenerateManifestTol, giqGenerateManifest)

	// Start the stdio server
	log.Printf("Starting GKE MCP Server")
	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
	}
}

func listClusters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", defaultProjectID)
	if projectID == "" {
		return mcp.NewToolResultError("project_id argument not set"), nil
	}
	location, _ := request.RequireString("location")
	if location == "" {
		location = "-"
	}

	c, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()

	req := &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectID, location),
	}
	resp, err := c.ListClusters(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(protojson.Format(resp)), nil
}

func getCluster(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", defaultProjectID)
	if projectID == "" {
		return mcp.NewToolResultError("project_id argument not set"), nil
	}
	location, err := request.RequireString("location")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	c, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	defer c.Close()

	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", projectID, location, name),
	}
	resp, err := c.GetCluster(ctx, req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(protojson.Format(resp)), nil
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

func getDefaultProject() string {
	out, err := exec.Command("gcloud", "config", "get", "core/project").Output()
	if err != nil {
		log.Printf("Failed to get default project: %v", err)
		return ""
	}
	projectID := strings.TrimSpace(string(out))
	log.Printf("Using default project ID: %s", projectID)
	return projectID
}
