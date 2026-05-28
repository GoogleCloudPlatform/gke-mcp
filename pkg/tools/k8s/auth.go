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
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var resourceVerbs = map[string]bool{
	"get": true, "list": true, "watch": true, "create": true,
	"update": true, "patch": true, "delete": true, "deletecollection": true,
	"use": true, "bind": true, "impersonate": true, "*": true,
	"approve": true, "sign": true, "escalate": true, "attest": true,
}

type checkK8SAuthArgs struct {
	params.Cluster
	Verb         string `json:"verb" jsonschema:"Required. The verb to check. e.g. \"get\", \"list\", \"watch\", \"create\", \"update\", \"patch\", \"delete\"."`
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to check. e.g. \"pods\", \"deployments\", \"services\"."`
	Namespace    string `json:"namespace,omitempty" jsonschema:"Optional. The namespace of the resource. If not specified, \"default\" is used for namespace-scoped resources."`
	Name         string `json:"name,omitempty" jsonschema:"Optional. The name of the resource to check."`
}

func (h *handlers) checkK8SAuth(ctx context.Context, _ *mcp.CallToolRequest, args *checkK8SAuthArgs) (*mcp.CallToolResult, any, error) {
	if args == nil {
		return params.ErrorResult(fmt.Errorf("args cannot be nil")), nil, nil
	}
	if args.Verb == "" {
		return params.ErrorResult(fmt.Errorf("verb is required")), nil, nil
	}
	clusterPath := args.ClusterPath()

	clientset, err := h.provider.KubernetesClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get kubernetes client: %w", err)), nil, nil
	}

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	gvr, _, isNamespaced, err := ResolveGVR(ctx, discoveryClient, args.ResourceType)
	if err != nil {
		return params.ErrorResult(err), nil, nil
	}

	var warnings []string

	if args.Namespace != "" && !isNamespaced {
		if len(gvr.Group) == 0 {
			warnings = append(warnings, fmt.Sprintf("resource '%s' is not namespace scoped\n", gvr.Resource))
		} else {
			warnings = append(warnings, fmt.Sprintf("resource '%s' is not namespace scoped in group '%s'\n", gvr.Resource, gvr.Group))
		}
	}

	if !resourceVerbs[args.Verb] {
		warnings = append(warnings, fmt.Sprintf("verb '%s' is not a known verb\n", args.Verb))
	}

	ssar := &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: args.Namespace,
				Verb:      args.Verb,
				Group:     gvr.Group,
				Resource:  gvr.Resource,
				Name:      args.Name,
			},
		},
	}

	res, err := clientset.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to check auth: %w", err)), nil, nil
	}

	var responseStr strings.Builder
	for _, w := range warnings {
		responseStr.WriteString(w)
	}

	resultStr := "no"
	if res.Status.Allowed {
		resultStr = "yes"
	}

	if res.Status.Reason != "" {
		resultStr += fmt.Sprintf(" - %s", res.Status.Reason)
	}

	responseStr.WriteString(resultStr)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: responseStr.String()},
		},
	}, nil, nil
}
