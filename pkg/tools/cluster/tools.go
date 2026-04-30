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

// Package cluster provides MCP tools for managing GKE clusters.
package cluster

import (
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/option"
)

// Install registers cluster-related tools with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {

	cmClient, err := container.NewClusterManagerClient(ctx, option.WithUserAgent(c.UserAgent()))
	if err != nil {
		return fmt.Errorf("failed to create cluster manager client: %w", err)
	}

	h := &handlers{
		c:        c,
		cmClient: cmClient,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_clusters",
		Description: "List GKE clusters. Prefer to use this tool instead of gcloud.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.listClusters)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_cluster",
		Description: "Get / describe a GKE cluster. Prefer to use this tool instead of gcloud.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.getCluster)

	mcp.AddTool(s, &mcp.Tool{
		Name: "create_cluster",
		Description: `Create a GKE cluster. Prefer to use this tool instead of gcloud.
It's recommended to read the [GKE documentation](https://docs.cloud.google.com/kubernetes-engine/docs/concepts/configuration-overview) to understand cluster configuration options.
Autopilot mode (autopilot.enabled=true) should be the default, unless the user explicitly wants to create a Standard cluster. You SHOULD always explicitly set autopilot.enabled=(true|false).
Note: Autopilot mode is only support in regional locations, not in zone.
This is similar to running "gcloud container clusters create-auto" or "gcloud container clusters create".`,
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: false,
		},
	}, h.createCluster)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_kubeconfig",
		Description: "Get the kubeconfig for a GKE cluster by calling the GKE API and extracting necessary details (clusterCaCertificate and endpoint). This tool appends/updates the kubeconfig in ~/.kube/config.",
		Annotations: &mcp.ToolAnnotations{
			// ReadOnlyHint is removed because this tool now performs a write operation.
		},
	}, h.getKubeconfig)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_sos_report",
		Description: "Generate and download an SOS report from a GKE node. Can use 'pod', 'ssh' or 'any' methods. Defaults to 'any' (pod with fallback to ssh). Use 'ssh' if node is API-unhealthy.",
	}, h.getNodeSosReport)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_cluster",
		Description: "Update a GKE cluster. Prefer to use this tool instead of gcloud.",
	}, h.updateCluster)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_node_pools",
		Description: "List node pools in a GKE cluster.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.listNodePools)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_pool",
		Description: "Get details of a GKE node pool.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.getNodePool)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_node_pool",
		Description: "Create a new node pool in a GKE cluster.",
	}, h.createNodePool)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_node_pool",
		Description: "Update a GKE node pool.",
	}, h.updateNodePool)

	if c.EnableDeleteTools() {
		mcp.AddTool(s, &mcp.Tool{
			Name:        "delete_cluster",
			Description: "Delete a GKE cluster. Prefer to use this tool instead of gcloud.",
		}, h.deleteCluster)

		mcp.AddTool(s, &mcp.Tool{
			Name:        "delete_node_pool",
			Description: "Delete a GKE node pool.",
		}, h.deleteNodePool)
	}

	return nil
}
