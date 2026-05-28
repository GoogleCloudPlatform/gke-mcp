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
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDescribeK8SResource(t *testing.T) {
	ctx := context.Background()

	// Create a fake unstructured pod
	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "my-pod",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "myapp",
				},
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "my-container",
						"image": "nginx",
					},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, pod)

	// Create fake clientset for events
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-event",
			Namespace: "default",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind:       "Pod",
			Name:       "my-pod",
			Namespace:  "default",
			APIVersion: "v1",
		},
		Message:       "Test event message",
		Reason:        "TestReason",
		Type:          "Normal",
		LastTimestamp: metav1.Time{Time: metav1.Now().Time},
	}
	fakeClientset := fake.NewSimpleClientset(event)

	// Mock discovery for ResolveGVR
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
		dynamicClient:    fakeDynamicClient,
		kubernetesClient: fakeClientset,
		discoveryClient:  fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &describeK8SResourceArgs{
		ResourceType: "pod",
		Name:         "my-pod",
		Namespace:    "default",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.describeK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("describeK8SResource failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("describeK8SResource returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "Name: my-pod") {
		t.Errorf("output does not contain 'Name: my-pod'")
	}
	if !strings.Contains(textContent.Text, "Kind: Pod") {
		t.Errorf("output does not contain 'Kind: Pod'")
	}
	if !strings.Contains(textContent.Text, "Spec:") {
		t.Errorf("output does not contain 'Spec:'")
	}
	if !strings.Contains(textContent.Text, "Events:") {
		t.Errorf("output does not contain 'Events:'")
	}
	if !strings.Contains(textContent.Text, "Test event message") {
		t.Errorf("output does not contain 'Test event message'")
	}
}

func TestDescribeK8SResource_NoResources(t *testing.T) {
	ctx := context.Background()

	scheme := runtime.NewScheme()
	gvrToKind := map[schema.GroupVersionResource]string{
		{Group: "", Version: "v1", Resource: "pods"}: "PodList",
	}
	fakeDynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToKind)

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
		dynamicClient:    fakeDynamicClient,
		kubernetesClient: fakeClientset,
		discoveryClient:  fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &describeK8SResourceArgs{
		ResourceType: "pod",
	}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.describeK8SResource(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("describeK8SResource failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("describeK8SResource returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	if !strings.Contains(textContent.Text, "No resources found.") {
		t.Errorf("output %q does not contain %q", textContent.Text, "No resources found.")
	}
}
