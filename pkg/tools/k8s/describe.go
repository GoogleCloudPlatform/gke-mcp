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
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type describeK8SResourceArgs struct {
	params.Cluster
	ResourceType  string `json:"resourceType" jsonschema:"Required. The type of the resource. e.g. \"pods\", \"deployments\", \"services\"."`
	Name          string `json:"name,omitempty" jsonschema:"Optional. The name of the resource. If not specified, all resources of the given type are described."`
	Namespace     string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, \"default\" is used for namespace-scoped resources."`
	LabelSelector string `json:"labelSelector,omitempty" jsonschema:"Optional. A label selector to filter resources."`
}

func (h *handlers) describeK8SResource(ctx context.Context, _ *mcp.CallToolRequest, args *describeK8SResourceArgs) (*mcp.CallToolResult, any, error) {
	if args == nil {
		return params.ErrorResult(fmt.Errorf("args cannot be nil")), nil, nil
	}
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	gvr, gvk, isNamespaced, err := ResolveGVR(ctx, discoveryClient, args.ResourceType)
	if err != nil {
		return params.ErrorResult(err), nil, nil
	}

	namespace := args.Namespace
	if isNamespaced && namespace == "" {
		namespace = "default"
	}

	dynamicClient, err := h.provider.DynamicClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get dynamic client: %w", err)), nil, nil
	}

	var resourceInterface dynamic.ResourceInterface
	if isNamespaced {
		resourceInterface = dynamicClient.Resource(gvr).Namespace(namespace)
	} else {
		resourceInterface = dynamicClient.Resource(gvr)
	}

	var resultBuilder strings.Builder

	if args.Name != "" {
		obj, err := resourceInterface.Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get resource: %w", err)), nil, nil
		}

		desc, err := h.describeObject(ctx, obj, gvk.Kind, isNamespaced, args.Cluster)
		if err != nil {
			return params.ErrorResult(err), nil, nil
		}
		resultBuilder.WriteString(desc)
	} else {
		list, err := resourceInterface.List(ctx, metav1.ListOptions{
			LabelSelector: args.LabelSelector,
		})
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to list resources: %w", err)), nil, nil
		}

		if len(list.Items) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No resources found."},
				},
			}, nil, nil
		}

		for i, item := range list.Items {
			desc, err := h.describeObject(ctx, &item, gvk.Kind, isNamespaced, args.Cluster)
			if err != nil {
				return params.ErrorResult(err), nil, nil
			}
			resultBuilder.WriteString(desc)
			if i < len(list.Items)-1 {
				resultBuilder.WriteString("\n---\n")
			}
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: resultBuilder.String()},
		},
	}, nil, nil
}

func (h *handlers) describeObject(ctx context.Context, obj *unstructured.Unstructured, kind string, isNamespaced bool, cluster params.Cluster) (string, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name: %s\n", obj.GetName())
	fmt.Fprintf(&sb, "Namespace: %s\n", obj.GetNamespace())
	fmt.Fprintf(&sb, "Kind: %s\n", kind)

	labels := obj.GetLabels()
	if len(labels) > 0 {
		sb.WriteString("Labels:\n")
		for k, v := range labels {
			fmt.Fprintf(&sb, "  %s=%s\n", k, v)
		}
	}

	annotations := obj.GetAnnotations()
	if len(annotations) > 0 {
		sb.WriteString("Annotations:\n")
		for k, v := range annotations {
			fmt.Fprintf(&sb, "  %s=%s\n", k, v)
		}
	}

	// Show spec and status as YAML
	spec, ok := obj.Object["spec"]
	if ok {
		specYAML, err := yaml.Marshal(spec)
		if err != nil {
			return "", fmt.Errorf("failed to marshal spec to YAML: %w", err)
		}
		sb.WriteString("Spec:\n")
		sb.WriteString(indent(string(specYAML), "  "))
	}

	status, ok := obj.Object["status"]
	if ok {
		statusYAML, err := yaml.Marshal(status)
		if err != nil {
			return "", fmt.Errorf("failed to marshal status to YAML: %w", err)
		}
		sb.WriteString("Status:\n")
		sb.WriteString(indent(string(statusYAML), "  "))
	}

	// Fetch events
	eventsResult, _, err := h.listK8SEvents(ctx, nil, &listK8SEventsArgs{
		Cluster:       cluster,
		Name:          obj.GetName(),
		Namespace:     obj.GetNamespace(),
		ResourceType:  kind,
		AllNamespaces: !isNamespaced,
	})

	if err == nil && !eventsResult.IsError && len(eventsResult.Content) > 0 {
		textContent, ok := eventsResult.Content[0].(*mcp.TextContent)
		if ok && textContent.Text != "" {
			sb.WriteString("\nEvents:\n")
			sb.WriteString(indent(textContent.Text, "  "))
		}
	}

	return sb.String(), nil
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}
