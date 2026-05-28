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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetK8SRolloutStatus_Table(t *testing.T) {
	tests := []struct {
		name         string
		args         *getK8SRolloutStatusArgs
		objects      []runtime.Object
		wantOutput   string
		discoveryRes []metav1.APIResource
	}{
		{
			name: "deployment success",
			args: &getK8SRolloutStatusArgs{ResourceType: "deployment", Name: "my-dep", Namespace: "default"},
			objects: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: "my-dep", Namespace: "default", Generation: 1},
					Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(1)},
					Status: appsv1.DeploymentStatus{
						ObservedGeneration: 1,
						UpdatedReplicas:    1,
						Replicas:           1,
						AvailableReplicas:  1,
					},
				},
			},
			wantOutput:   "deployment \"my-dep\" successfully rolled out",
			discoveryRes: []metav1.APIResource{{Name: "deployments", Namespaced: true, Kind: "Deployment"}},
		},
		{
			name: "deployment pending",
			args: &getK8SRolloutStatusArgs{ResourceType: "deployment", Name: "my-dep", Namespace: "default"},
			objects: []runtime.Object{
				&appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{Name: "my-dep", Namespace: "default", Generation: 2},
					Spec:       appsv1.DeploymentSpec{Replicas: int32Ptr(2)},
					Status: appsv1.DeploymentStatus{
						ObservedGeneration: 2,
						UpdatedReplicas:    1,
						Replicas:           2,
						AvailableReplicas:  1,
					},
				},
			},
			wantOutput:   "waiting for rollout to finish: 1 out of 2 new replicas have been updated",
			discoveryRes: []metav1.APIResource{{Name: "deployments", Namespaced: true, Kind: "Deployment"}},
		},
		{
			name: "daemonset success",
			args: &getK8SRolloutStatusArgs{ResourceType: "daemonset", Name: "my-ds", Namespace: "default"},
			objects: []runtime.Object{
				&appsv1.DaemonSet{
					ObjectMeta: metav1.ObjectMeta{Name: "my-ds", Namespace: "default", Generation: 1},
					Status: appsv1.DaemonSetStatus{
						ObservedGeneration:     1,
						UpdatedNumberScheduled: 1,
						DesiredNumberScheduled: 1,
						NumberAvailable:        1,
					},
				},
			},
			wantOutput:   "daemonset \"my-ds\" successfully rolled out",
			discoveryRes: []metav1.APIResource{{Name: "daemonsets", Namespaced: true, Kind: "DaemonSet"}},
		},
		{
			name: "statefulset success",
			args: &getK8SRolloutStatusArgs{ResourceType: "statefulset", Name: "my-ss", Namespace: "default"},
			objects: []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "my-ss", Namespace: "default", Generation: 1},
					Spec:       appsv1.StatefulSetSpec{Replicas: int32Ptr(1)},
					Status: appsv1.StatefulSetStatus{
						ObservedGeneration: 1,
						ReadyReplicas:      1,
						UpdatedReplicas:    1,
					},
				},
			},
			wantOutput:   "statefulset \"my-ss\" successfully rolled out",
			discoveryRes: []metav1.APIResource{{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet"}},
		},
		{
			name: "statefulset partitioned success",
			args: &getK8SRolloutStatusArgs{ResourceType: "statefulset", Name: "my-ss-part", Namespace: "default"},
			objects: []runtime.Object{
				&appsv1.StatefulSet{
					ObjectMeta: metav1.ObjectMeta{Name: "my-ss-part", Namespace: "default", Generation: 1},
					Spec: appsv1.StatefulSetSpec{
						Replicas: int32Ptr(2),
						UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
							Type: appsv1.RollingUpdateStatefulSetStrategyType,
							RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
								Partition: int32Ptr(1),
							},
						},
					},
					Status: appsv1.StatefulSetStatus{
						ObservedGeneration: 1,
						ReadyReplicas:      2,
						UpdatedReplicas:    1,
					},
				},
			},
			wantOutput:   "statefulset \"my-ss-part\" successfully rolled out",
			discoveryRes: []metav1.APIResource{{Name: "statefulsets", Namespaced: true, Kind: "StatefulSet"}},
		},
		{
			name: "unsupported kind",
			args: &getK8SRolloutStatusArgs{ResourceType: "pod", Name: "my-pod", Namespace: "default"},
			objects: []runtime.Object{
				&appsv1.Deployment{ // Just need some object to avoid empty client errors if any
					ObjectMeta: metav1.ObjectMeta{Name: "my-dep", Namespace: "default"},
				},
			},
			wantOutput:   "rollout status not supported for resource of kind \"Pod\"",
			discoveryRes: []metav1.APIResource{{Name: "pods", Namespaced: true, Kind: "Pod"}},
		},
		{
			name:         "missing namespace",
			args:         &getK8SRolloutStatusArgs{ResourceType: "deployment", Name: "my-dep"},
			wantOutput:   "namespace is required",
			discoveryRes: []metav1.APIResource{{Name: "deployments", Namespaced: true, Kind: "Deployment"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			fakeClientset := fake.NewSimpleClientset(tt.objects...)

			fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
			fakeDiscovery.Resources = []*metav1.APIResourceList{
				{
					GroupVersion: "apps/v1",
					APIResources: tt.discoveryRes,
				},
				{
					GroupVersion: "v1",
					APIResources: tt.discoveryRes, // For pods if tested
				},
			}

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

			result, _, err := h.getK8SRolloutStatus(ctx, &mcp.CallToolRequest{}, tt.args)
			if err != nil {
				t.Fatalf("getK8SRolloutStatus failed: %v", err)
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

func int32Ptr(i int32) *int32 { return &i }
