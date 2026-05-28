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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPatchK8SResource_Table(t *testing.T) {
	tests := []struct {
		name         string
		args         *patchK8SResourceArgs
		initialObj   *unstructured.Unstructured
		wantErr      bool
		wantOutput   string
		wantErrorStr string
	}{
		{
			name: "successful patch",
			args: &patchK8SResourceArgs{ResourceType: "pod", Name: "my-pod", Namespace: "default", Patch: `{"metadata":{"labels":{"new-label":"value"}}}`, PatchType: "merge"},
			initialObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "default",
					},
				},
			},
			wantErr:    false,
			wantOutput: "new-label: value",
		},
		{
			name: "resource not found",
			args: &patchK8SResourceArgs{ResourceType: "pod", Name: "non-existent-pod", Namespace: "default", Patch: `{"metadata":{"labels":{"new-label":"value"}}}`},
			initialObj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "default",
					},
				},
			},
			wantErr:      true,
			wantErrorStr: "failed to patch resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()

			var fakeDynamicClient *dynamicfake.FakeDynamicClient
			if tt.initialObj != nil {
				fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, tt.initialObj)
			} else {
				fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme)
			}

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

			tt.args.ProjectID = "p"
			tt.args.Location = "l"
			tt.args.ClusterName = "c"

			result, _, err := h.patchK8SResource(ctx, &mcp.CallToolRequest{}, tt.args)

			if tt.wantErr {
				if err != nil {
					t.Fatalf("expected error result, but got error: %v", err)
				}
				if !result.IsError {
					t.Fatalf("expected error result, but got success")
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					t.Fatalf("result.Content[0] is not TextContent")
				}
				if !strings.Contains(textContent.Text, tt.wantErrorStr) {
					t.Errorf("error output %q does not contain %q", textContent.Text, tt.wantErrorStr)
				}
			} else {
				if err != nil {
					t.Fatalf("patchK8SResource failed: %v", err)
				}
				if result.IsError {
					t.Fatalf("patchK8SResource returned error result: %v", result.Content[0])
				}
				textContent, ok := result.Content[0].(*mcp.TextContent)
				if !ok {
					t.Fatalf("result.Content[0] is not TextContent")
				}
				if !strings.Contains(textContent.Text, tt.wantOutput) {
					t.Errorf("output %q does not contain %q", textContent.Text, tt.wantOutput)
				}
			}
		})
	}
}
