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
	"encoding/json"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListK8SAPIResources(t *testing.T) {
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
		{
			GroupVersion: "apps/v1",
			APIResources: []metav1.APIResource{
				{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			},
		},
		{
			GroupVersion: "custom.example.com/v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Namespaced: true, Kind: "CustomPod"},
			},
		},
	}

	mockProvider := &mockClientProvider{
		discoveryClient: fakeDiscovery,
	}

	h := &handlers{
		c:        &config.Config{},
		provider: mockProvider,
	}

	args := &listK8SAPIResourcesArgs{}
	args.ProjectID = "p"
	args.Location = "l"
	args.ClusterName = "c"

	result, _, err := h.listK8SAPIResources(ctx, &mcp.CallToolRequest{}, args)
	if err != nil {
		t.Fatalf("listK8SAPIResources failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("listK8SAPIResources returned error result: %v", result.Content[0])
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("result.Content[0] is not TextContent")
	}

	var resources []APIGroupDiscovery
	if err := json.Unmarshal([]byte(textContent.Text), &resources); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(resources) != 3 {
		t.Fatalf("len(resources) = %d, want 3", len(resources))
	}

	// Verify content (assuming sorted by name)
	// Expected order: deployments, pods, pods

	if resources[0].Name != "deployments" {
		t.Errorf("resources[0].Name = %s, want deployments", resources[0].Name)
	}

	// We expect pods to be at index 1 and 2.
	foundCorePods := false
	foundCustomPods := false

	for i := 1; i <= 2; i++ {
		r := resources[i]
		if r.Name != "pods" {
			t.Errorf("resources[%d].Name = %s, want pods", i, r.Name)
			continue
		}
		if len(r.Versions) != 1 {
			t.Errorf("resources[%d].Versions length = %d, want 1", i, len(r.Versions))
			continue
		}
		if r.Versions[0] == "v1" {
			foundCorePods = true
		}
		if r.Versions[0] == "custom.example.com/v1" {
			foundCustomPods = true
		}
	}

	if !foundCorePods {
		t.Errorf("core pods (v1) not found at index 1 or 2")
	}
	if !foundCustomPods {
		t.Errorf("custom pods (custom.example.com/v1) not found at index 1 or 2")
	}
}
