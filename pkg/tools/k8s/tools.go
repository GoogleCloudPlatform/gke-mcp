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

// Package k8s provides MCP tools for interacting with Kubernetes resources.
package k8s

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type handlers struct {
	c        *config.Config
	provider Provider
}

// Install registers Kubernetes-related tools with the MCP server.
func Install(_ context.Context, s *mcp.Server, c *config.Config) error {
	h := &handlers{
		c:        c,
		provider: NewClientProvider(),
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_resource",
		Description: "Gets one or more Kubernetes resources from a cluster. Resources can be filtered by type, name, namespace, and label selectors. Returns the resources in YAML format. This is similar to running `kubectl get`.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.getK8SResource)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_k8s_events",
		Description: "Retrieves events from a Kubernetes cluster. This is similar to running `kubectl events`.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.listK8SEvents)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_k8s_version",
		Description: "Retrieves the Kubernetes server version for a given cluster. This is similar to running kubectl version.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.getK8SVersion)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "apply_k8s_manifest",
		Description: "Applies a Kubernetes manifest to a cluster using server-side apply. This is similar to running `kubectl apply --server-side`.",
	}, h.applyK8SManifest)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_k8s_resource",
		Description: "Deletes a Kubernetes resource from a cluster. This is similar to running `kubectl delete`.",
	}, h.deleteK8SResource)

	return nil
}
