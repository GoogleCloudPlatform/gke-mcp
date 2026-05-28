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
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestCheckK8SAuth_Table(t *testing.T) {
	tests := []struct {
		name         string
		args         *checkK8SAuthArgs
		allowed      bool
		reason       string
		wantOutput   string
		discoveryRes []metav1.APIResource
	}{
		{
			name:         "allowed",
			args:         &checkK8SAuthArgs{Verb: "get", ResourceType: "pod", Namespace: "default"},
			allowed:      true,
			reason:       "Allowed by test",
			wantOutput:   "yes - Allowed by test",
			discoveryRes: []metav1.APIResource{{Name: "pods", Namespaced: true, Kind: "Pod"}},
		},
		{
			name:         "denied",
			args:         &checkK8SAuthArgs{Verb: "create", ResourceType: "pod", Namespace: "default"},
			allowed:      false,
			reason:       "Denied by test",
			wantOutput:   "no - Denied by test",
			discoveryRes: []metav1.APIResource{{Name: "pods", Namespaced: true, Kind: "Pod"}},
		},
		{
			name:         "unknown verb warning",
			args:         &checkK8SAuthArgs{Verb: "invalid_verb", ResourceType: "pod", Namespace: "default"},
			allowed:      true,
			wantOutput:   "verb 'invalid_verb' is not a known verb\nyes",
			discoveryRes: []metav1.APIResource{{Name: "pods", Namespaced: true, Kind: "Pod"}},
		},
		{
			name:         "cluster scoped resource with namespace warning",
			args:         &checkK8SAuthArgs{Verb: "get", ResourceType: "node", Namespace: "default"},
			allowed:      true,
			wantOutput:   "resource 'nodes' is not namespace scoped\nyes",
			discoveryRes: []metav1.APIResource{{Name: "nodes", Namespaced: false, Kind: "Node"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fakeClientset := fake.NewSimpleClientset()

			fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
			fakeDiscovery.Resources = []*metav1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: tt.discoveryRes,
				},
			}

			fakeClientset.PrependReactor("create", "selfsubjectaccessreviews", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
				createAction := action.(k8stesting.CreateAction)
				obj := createAction.GetObject().(*authorizationv1.SelfSubjectAccessReview)
				obj.Status.Allowed = tt.allowed
				obj.Status.Reason = tt.reason
				return true, obj, nil
			})

			mockProvider := &mockClientProvider{
				kubernetesClient: fakeClientset,
				discoveryClient:  fakeDiscovery,
			}

			h := &handlers{
				c:        &config.Config{},
				provider: mockProvider,
			}

			tt.args.ProjectID = "p"
			tt.args.Location = "l"
			tt.args.ClusterName = "c"

			result, _, err := h.checkK8SAuth(ctx, &mcp.CallToolRequest{}, tt.args)
			if err != nil {
				t.Fatalf("checkK8SAuth failed: %v", err)
			}

			textContent, ok := result.Content[0].(*mcp.TextContent)
			if !ok {
				t.Fatalf("result.Content[0] is not TextContent")
			}

			if !strings.Contains(textContent.Text, tt.wantOutput) {
				t.Errorf("output %q does not contain %q", textContent.Text, tt.wantOutput)
			}
		})
	}
}
