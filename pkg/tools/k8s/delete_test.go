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
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeleteK8SResource(t *testing.T) {
	ctx := context.Background()

	// Create a fake unstructured pod
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "my-pod",
				"namespace": "default",
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, pod)

	// We need to mock discovery for ResolveGVR
	fakeClientset := fake.NewSimpleClientset()
	fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
	}

	mockProvider := &mockClientProvider{
		dynamicClient:   fakeDynamicClient,
		discoveryClient: fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &deleteK8SResourceArgs{
		ResourceType: "pod",
		Name:         "my-pod",
		Namespace:    "default",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.deleteK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("deleteK8SResource failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("deleteK8SResource returned error result: %v", result.Content[0])
	}

	if len(result.Content) != 1 {
		t.Fatalf("len(result.Content) = %d, want 1", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "deleted") {
		t.Errorf("output does not contain 'deleted'")
	}
}

func TestDeleteK8SResource_DryRun(t *testing.T) {
	ctx := context.Background()

	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "my-pod",
				"namespace": "default",
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, pod)

	fakeClientset := fake.NewSimpleClientset()
	fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
	}

	mockProvider := &mockClientProvider{
		dynamicClient:   fakeDynamicClient,
		discoveryClient: fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &deleteK8SResourceArgs{
		ResourceType: "pod",
		Name:         "my-pod",
		Namespace:    "default",
		DryRun:       true,
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.deleteK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("deleteK8SResource failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("deleteK8SResource returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "(dry-run)") {
		t.Errorf("output does not contain '(dry-run)'")
	}
}

func TestDeleteK8SResource_InvalidCascade(t *testing.T) {
	ctx := context.Background()

	fakeClientset := fake.NewSimpleClientset()
	fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "Pod"},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)

	h := &handlers{
		c: &config.Config{},
		provider: &mockClientProvider{
			discoveryClient: fakeDiscovery,
			dynamicClient:   fakeDynamicClient,
		},
	}

	args := &deleteK8SResourceArgs{
		ResourceType: "pod",
		Name:         "my-pod",
		Namespace:    "default",
		Cascade:      "invalid",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.deleteK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("deleteK8SResource failed: %v", err)
	}

	if !result.IsError {
		t.Fatalf("deleteK8SResource expected error result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "invalid cascade policy") {
		t.Errorf("output does not contain 'invalid cascade policy'")
	}
}

func TestDeleteK8SResource_ClusterScoped(t *testing.T) {
	ctx := context.Background()

	node := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Node",
			"metadata": map[string]interface{}{
				"name": "my-node",
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, node)

	fakeClientset := fake.NewSimpleClientset()
	fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
	fakeDiscovery.Resources = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "nodes", Namespaced: false, Kind: "Node"},
			},
		},
	}

	mockProvider := &mockClientProvider{
		dynamicClient:   fakeDynamicClient,
		discoveryClient: fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &deleteK8SResourceArgs{
		ResourceType: "node",
		Name:         "my-node",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.deleteK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("deleteK8SResource failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("deleteK8SResource returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	expectedMsg := "Resource my-node (kind: node) deleted"
	if textContent.Text != expectedMsg {
		t.Errorf("output = %q, want %q", textContent.Text, expectedMsg)
	}
}
