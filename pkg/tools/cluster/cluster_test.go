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

package cluster

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func TestListClustersArgs_Fields(t *testing.T) {
	args := listClustersArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ReadMask = "name,status"

	if args.LocationPath() != "projects/test-project/locations/us-central1" {
		t.Errorf("LocationPath() = %s, want projects/test-project/locations/us-central1", args.LocationPath())
	}
	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
}

func TestGetClustersArgs_Fields(t *testing.T) {
	args := getClustersArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.ReadMask = "name,status"

	if args.ClusterPath() != "projects/test-project/locations/us-central1/clusters/my-cluster" {
		t.Errorf("ClusterPath() = %s, want projects/test-project/locations/us-central1/clusters/my-cluster", args.ClusterPath())
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
}

func TestCreateClustersArgs_Fields(t *testing.T) {
	args := createClustersArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.Cluster = `{"name": "my-cluster"}`

	if args.LocationPath() != "projects/test-project/locations/us-central1" {
		t.Errorf("LocationPath() = %s, want projects/test-project/locations/us-central1", args.LocationPath())
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.Cluster != `{"name": "my-cluster"}` {
		t.Errorf("Cluster = %s, want {\"name\": \"my-cluster\"}", args.Cluster)
	}
}

func TestGetKubeconfigArgs_Fields(t *testing.T) {
	var args getKubeconfigArgs
	args.ProjectID = "test-project"
	args.Location = "us-west1"
	args.ClusterName = "my-cluster"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-west1" {
		t.Errorf("Location = %s, want us-west1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
}

func TestGetNodeSosReportArgs_Fields(t *testing.T) {
	args := getNodeSosReportArgs{
		Node:           "my-node",
		Destination:    "/tmp/sos",
		Method:         "pod",
		TimeoutSeconds: 300,
	}

	if args.Node != "my-node" {
		t.Errorf("Node = %s, want my-node", args.Node)
	}
	if args.Destination != "/tmp/sos" {
		t.Errorf("Destination = %s, want /tmp/sos", args.Destination)
	}
	if args.Method != "pod" {
		t.Errorf("Method = %s, want pod", args.Method)
	}
	if args.TimeoutSeconds != 300 {
		t.Errorf("TimeoutSeconds = %d, want 300", args.TimeoutSeconds)
	}
}

func TestListClustersArgs_Empty(t *testing.T) {
	args := listClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.ReadMask != "" {
		t.Errorf("Expected empty ReadMask, got %s", args.ReadMask)
	}
}

func TestGetClustersArgs_Empty(t *testing.T) {
	args := getClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.ClusterName != "" {
		t.Errorf("Expected empty ClusterName, got %s", args.ClusterName)
	}
	if args.ReadMask != "" {
		t.Errorf("Expected empty ReadMask, got %s", args.ReadMask)
	}
}

func TestCreateClustersArgs_Empty(t *testing.T) {
	args := createClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.Cluster != "" {
		t.Errorf("Expected empty Cluster, got %s", args.Cluster)
	}
}

func TestUpdateClusterArgs_Fields(t *testing.T) {
	args := updateClusterArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.Update = `{"description": "new description"}`

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.Update != `{"description": "new description"}` {
		t.Errorf("Update = %s, want {\"description\": \"new description\"}", args.Update)
	}
}

func TestDeleteClusterArgs_Fields(t *testing.T) {
	args := deleteClusterArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.DeletionPolicy = "FORCE"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.DeletionPolicy != "FORCE" {
		t.Errorf("DeletionPolicy = %s, want FORCE", args.DeletionPolicy)
	}
}

type mockK8sProvider struct {
	dynamicClient   dynamic.Interface
	discoveryClient discovery.DiscoveryInterface
	err             error
}

func (m *mockK8sProvider) RESTConfig(_ context.Context, _ string) (*rest.Config, error) {
	return nil, m.err
}

func (m *mockK8sProvider) DynamicClient(_ context.Context, _ string) (dynamic.Interface, error) {
	return m.dynamicClient, m.err
}

func (m *mockK8sProvider) DynamicClientWithHeaders(_ context.Context, _ string, _, _ string) (dynamic.Interface, error) {
	return m.dynamicClient, m.err
}

func (m *mockK8sProvider) DiscoveryClient(_ context.Context, _ string) (discovery.DiscoveryInterface, error) {
	return m.discoveryClient, m.err
}

func (m *mockK8sProvider) KubernetesClient(_ context.Context, _ string) (kubernetes.Interface, error) {
	return nil, m.err
}

func TestVerifyClusterUnused(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		objects []runtime.Object
		wantErr bool
	}{
		{
			name: "failed_due_to_load_balancer",
			objects: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]interface{}{
							"name":      "lb-svc",
							"namespace": "default",
						},
						"spec": map[string]interface{}{
							"type": "LoadBalancer",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed_due_to_bound_pvc",
			objects: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "PersistentVolumeClaim",
						"metadata": map[string]interface{}{
							"name":      "my-pvc",
							"namespace": "default",
						},
						"status": map[string]interface{}{
							"phase": "Bound",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed_due_to_running_pod",
			objects: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "app-pod",
							"namespace": "default",
						},
						"status": map[string]interface{}{
							"phase": "Running",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "failed_due_to_active_ingress",
			objects: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "networking.k8s.io/v1",
						"kind":       "Ingress",
						"metadata": map[string]interface{}{
							"name":      "my-ingress",
							"namespace": "default",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "succeeds_with_kube_system_pod",
			objects: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Pod",
						"metadata": map[string]interface{}{
							"name":      "sys-pod",
							"namespace": "kube-system",
						},
						"status": map[string]interface{}{
							"phase": "Running",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "succeeds_when_unused",
			objects: []runtime.Object{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			gvrToKind := map[schema.GroupVersionResource]string{
				{Group: "", Version: "v1", Resource: "services"}:                          "ServiceList",
				{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}:            "PersistentVolumeClaimList",
				{Group: "", Version: "v1", Resource: "pods"}:                              "PodList",
				{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}:        "IngressList",
				{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "gateways"}: "GatewayList",
			}
			fakeDynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToKind, tt.objects...)

			fakeClientset := fake.NewSimpleClientset()
			fakeDiscovery := fakeClientset.Discovery().(*fakediscovery.FakeDiscovery)
			fakeDiscovery.Resources = []*metav1.APIResourceList{
				{
					GroupVersion: "v1",
					APIResources: []metav1.APIResource{
						{Name: "pods", Namespaced: true, Kind: "Pod"},
						{Name: "services", Namespaced: true, Kind: "Service"},
						{Name: "persistentvolumeclaims", Namespaced: true, Kind: "PersistentVolumeClaim"},
					},
				},
				{
					GroupVersion: "networking.k8s.io/v1",
					APIResources: []metav1.APIResource{
						{Name: "ingresses", Namespaced: true, Kind: "Ingress"},
					},
				},
			}

			h := &handlers{
				k8sProvider: &mockK8sProvider{
					dynamicClient:   fakeDynamicClient,
					discoveryClient: fakeDiscovery,
				},
			}

			err := h.verifyClusterUnused(ctx, "projects/my-project/locations/us-central1/clusters/my-cluster")
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyClusterUnused() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
