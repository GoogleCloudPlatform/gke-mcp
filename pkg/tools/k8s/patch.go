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

package k8s

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type patchK8SResourceArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to patch. e.g. \"pods\", \"deployments\", \"services\"."`
	Name         string `json:"name" jsonschema:"Required. The name of the resource to patch."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, \"default\" is used."`
	Patch        string `json:"patch" jsonschema:"Required. The patch to apply in JSON format."`
	PatchType    string `json:"patchType,omitempty" jsonschema:"Optional. The type of patch. One of: (json, merge, strategic). Defaults to strategic if not specified."`
}

func (h *handlers) patchK8SResource(ctx context.Context, _ *mcp.CallToolRequest, args *patchK8SResourceArgs) (*mcp.CallToolResult, any, error) {
	if args == nil {
		return params.ErrorResult(fmt.Errorf("args cannot be nil")), nil, nil
	}
	if args.ProjectID == "" {
		return params.ErrorResult(fmt.Errorf("projectId is required")), nil, nil
	}
	if args.Location == "" {
		return params.ErrorResult(fmt.Errorf("location is required")), nil, nil
	}
	if args.ClusterName == "" {
		return params.ErrorResult(fmt.Errorf("clusterName is required")), nil, nil
	}
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

	namespace := args.Namespace
	if isNamespaced && namespace == "" {
		namespace = "default"
	}

	var pt types.PatchType
	switch args.PatchType {
	case "json":
		pt = types.JSONPatchType
	case "merge":
		pt = types.MergePatchType
	case "strategic", "":
		pt = types.StrategicMergePatchType
	default:
		return params.ErrorResult(fmt.Errorf("unknown patch type: %s", args.PatchType)), nil, nil
	}

	var resourceInterface dynamic.ResourceInterface
	if isNamespaced {
		resourceInterface = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resourceInterface = dynamicClient.Resource(gvr)
	}
	patchBytes := []byte(args.Patch)
	if pt == types.StrategicMergePatchType || pt == types.MergePatchType {
		var err error
		patchBytes, err = yaml.YAMLToJSON(patchBytes)
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to convert patch to JSON: %w", err)), nil, nil
		}
	}

	patchedObj, err := resourceInterface.Patch(ctx, args.Name, pt, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to patch resource: %w", err)), nil, nil
	}

	data, err := yaml.Marshal(patchedObj)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to marshal patched resource to YAML: %w", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}
