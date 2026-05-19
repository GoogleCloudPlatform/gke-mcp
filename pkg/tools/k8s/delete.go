// Copyright 2026 Google LLC
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

package k8s

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type deleteK8SResourceArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to delete. Kubernetes resource/kind name in singular form, lower case. e.g. \"pod\", \"deployment\", \"service\"."`
	Name         string `json:"name" jsonschema:"Required. The name of the resource to delete."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, \"default\" is used."`
	Cascade      string `json:"cascade,omitempty" jsonschema:"Optional. The cascading deletion policy to use. If not specified, 'background' is used. Valid values are 'background', 'foreground', and 'orphan'."`
	DryRun       bool   `json:"dryRun,omitempty" jsonschema:"Optional. If true, run in dry-run mode."`
}

func (h *handlers) deleteK8SResource(ctx context.Context, _ *mcp.CallToolRequest, args *deleteK8SResourceArgs) (*mcp.CallToolResult, any, error) {
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	gvr, _, isNamespaced, err := ResolveGVR(ctx, discoveryClient, args.ResourceType)
	if err != nil {
		return params.ErrorResult(err), nil, nil
	}

	dynamicClient, err := h.provider.DynamicClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get dynamic client: %w", err)), nil, nil
	}

	var resourceInterface dynamic.ResourceInterface
	var resourceDesc string
	if isNamespaced {
		ns := args.Namespace
		if ns == "" {
			ns = "default"
		}
		resourceInterface = dynamicClient.Resource(gvr).Namespace(ns)
		resourceDesc = fmt.Sprintf("%s/%s", ns, args.Name)
	} else {
		resourceInterface = dynamicClient.Resource(gvr)
		resourceDesc = args.Name
	}

	deleteOptions := metav1.DeleteOptions{}
	if args.Cascade != "" {
		var propagationPolicy metav1.DeletionPropagation
		switch args.Cascade {
		case "background":
			propagationPolicy = metav1.DeletePropagationBackground
		case "foreground":
			propagationPolicy = metav1.DeletePropagationForeground
		case "orphan":
			propagationPolicy = metav1.DeletePropagationOrphan
		default:
			return params.ErrorResult(fmt.Errorf("invalid cascade policy: %s", args.Cascade)), nil, nil
		}
		deleteOptions.PropagationPolicy = &propagationPolicy
	}

	if args.DryRun {
		deleteOptions.DryRun = []string{"All"}
	}

	err = resourceInterface.Delete(ctx, args.Name, deleteOptions)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to delete resource: %w", err)), nil, nil
	}

	result := fmt.Sprintf("Resource %s (kind: %s) deleted", resourceDesc, args.ResourceType)
	if args.DryRun {
		result += " (dry-run)"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}
