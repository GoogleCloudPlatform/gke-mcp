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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
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

func TestResolveSosDestinationDir(t *testing.T) {
	baseDir := t.TempDir()

	t.Run("empty destination creates temp directory", func(t *testing.T) {
		tmpRoot := t.TempDir()
		t.Setenv("TMPDIR", tmpRoot)

		got, cleanupRoot, err := resolveSosDestinationDir(baseDir, "")
		if err != nil {
			t.Fatalf("resolveSosDestinationDir(\"\") unexpected error: %v", err)
		}
		if cleanupRoot != got {
			t.Errorf("cleanupRoot = %q, want %q", cleanupRoot, got)
		}
		rel, err := filepath.Rel(tmpRoot, got)
		if err != nil || !filepath.IsLocal(rel) || rel == "." {
			t.Errorf("temp dir %q is not inside TMPDIR %q", got, tmpRoot)
		}
		info, err := os.Stat(got)
		if err != nil || !info.IsDir() {
			t.Errorf("expected temp dir %q to exist as directory, err = %v", got, err)
		}
		if perm := info.Mode().Perm(); perm != 0o700 {
			t.Errorf("temp dir perm = %o, want 0700", perm)
		}
	})

	t.Run("cleanupRoot tracks only newly created prefix", func(t *testing.T) {
		existingDir := filepath.Join(baseDir, "already-here")
		if err := os.Mkdir(existingDir, 0o750); err != nil {
			t.Fatalf("Mkdir(%q) failed: %v", existingDir, err)
		}

		got, cleanupRoot, err := resolveSosDestinationDir(baseDir, "already-here")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != existingDir || cleanupRoot != "" {
			t.Errorf("got (%q, %q), want (%q, \"\")", got, cleanupRoot, existingDir)
		}

		got, cleanupRoot, err = resolveSosDestinationDir(baseDir, filepath.Join("already-here", "new-child", "grandchild"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantCleanup := filepath.Join(existingDir, "new-child")
		if cleanupRoot != wantCleanup {
			t.Errorf("cleanupRoot = %q, want %q (got=%q)", cleanupRoot, wantCleanup, got)
		}
	})

	tests := []struct {
		name    string
		rawDest string
		wantRel string
		wantErr string
	}{
		{
			name:    "current directory dot",
			rawDest: ".",
			wantRel: ".",
		},
		{
			name:    "current directory dot slash",
			rawDest: "./",
			wantRel: ".",
		},
		{
			name:    "relative subdirectory",
			rawDest: "sos-reports",
			wantRel: "sos-reports",
		},
		{
			name:    "nested relative path",
			rawDest: filepath.Join("a", "b", "c"),
			wantRel: filepath.Join("a", "b", "c"),
		},
		{
			name:    "internal traversal staying inside working directory",
			rawDest: filepath.Join("a", "..", "b"),
			wantRel: "b",
		},
		{
			name:    "whitespace only",
			rawDest: "   ",
			wantErr: "cannot be whitespace",
		},
		{
			name:    "nul byte",
			rawDest: "foo\x00bar",
			wantErr: "invalid control character",
		},
		{
			name:    "newline control char",
			rawDest: "a\nb",
			wantErr: "invalid control character",
		},
		{
			name:    "carriage return control char",
			rawDest: "a\rb",
			wantErr: "invalid control character",
		},
		{
			name:    "leading single dash",
			rawDest: "-foo",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "leading double dash flag",
			rawDest: "--foo",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "leading dash hidden behind dot slash",
			rawDest: "./-foo",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "leading dash hidden behind traversal",
			rawDest: filepath.Join("a", "..", "-foo"),
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "absolute path",
			rawDest: "/etc/x",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "root path",
			rawDest: "/",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "parent directory",
			rawDest: "..",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "parent directory escape",
			rawDest: "../outside",
			wantErr: "must be a relative path under the current directory",
		},
		{
			name:    "nested parent directory escape",
			rawDest: filepath.Join("a", "..", "..", "outside"),
			wantErr: "must be a relative path under the current directory",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, cleanupRoot, err := resolveSosDestinationDir(baseDir, tc.rawDest)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("resolveSosDestinationDir(%q) error = nil, want substring %q", tc.rawDest, tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("resolveSosDestinationDir(%q) error = %q, want substring %q", tc.rawDest, err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveSosDestinationDir(%q) unexpected error: %v", tc.rawDest, err)
			}
			want := filepath.Join(baseDir, tc.wantRel)
			if got != want {
				t.Errorf("resolveSosDestinationDir(%q) = %q, want %q", tc.rawDest, got, want)
			}
			if tc.rawDest == "." || tc.rawDest == "./" {
				if cleanupRoot != "" {
					t.Errorf("cleanupRoot = %q, want \"\" for %q", cleanupRoot, tc.rawDest)
				}
			} else if cleanupRoot == "" {
				t.Errorf("cleanupRoot = \"\", expected non-empty for new dir %q", tc.rawDest)
			}
			info, statErr := os.Stat(got)
			if statErr != nil || !info.IsDir() {
				t.Errorf("expected directory %q to exist, statErr = %v", got, statErr)
			} else if tc.rawDest != "." && tc.rawDest != "./" && (info.Mode().Perm()&0o027) != 0 {
				t.Errorf("directory %q permissions = %o, want no group-write or other permissions", got, info.Mode().Perm())
			}
		})
	}
}

func TestResolveSosDestinationDir_SymlinksAndFiles(t *testing.T) {
	baseDir := t.TempDir()
	outsideDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(outsideDir, "existing"), 0o750); err != nil {
		t.Fatalf("mkdir outside existing: %v", err)
	}
	if err := os.Mkdir(filepath.Join(baseDir, "real-sub"), 0o750); err != nil {
		t.Fatalf("mkdir real-sub: %v", err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(baseDir, "escape-link")); err != nil {
		t.Fatalf("symlink escape-link: %v", err)
	}
	if err := os.Symlink(filepath.Join(outsideDir, "does-not-exist"), filepath.Join(baseDir, "dangling-link")); err != nil {
		t.Fatalf("symlink dangling-link: %v", err)
	}
	if err := os.Symlink(filepath.Join(baseDir, "real-sub"), filepath.Join(baseDir, "internal-link")); err != nil {
		t.Fatalf("symlink internal-link: %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseDir, "regular-file"), []byte("x"), 0o600); err != nil {
		t.Fatalf("write regular-file: %v", err)
	}

	t.Run("symlink escaping baseDir with non-existent child", func(t *testing.T) {
		_, _, err := resolveSosDestinationDir(baseDir, filepath.Join("escape-link", "new-dir"))
		if err == nil || !strings.Contains(err.Error(), "must resolve within the current directory") {
			t.Fatalf("expected symlink escape error, got: %v", err)
		}
		if _, statErr := os.Lstat(filepath.Join(outsideDir, "new-dir")); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("directory should not have been created outside baseDir, statErr = %v", statErr)
		}
	})

	t.Run("symlink escaping baseDir with existing child", func(t *testing.T) {
		_, _, err := resolveSosDestinationDir(baseDir, filepath.Join("escape-link", "existing"))
		if err == nil || !strings.Contains(err.Error(), "must resolve within the current directory") {
			t.Fatalf("expected symlink escape error, got: %v", err)
		}
	})

	t.Run("dangling symlink rejected", func(t *testing.T) {
		_, _, err := resolveSosDestinationDir(baseDir, "dangling-link")
		if err == nil || !strings.Contains(err.Error(), "unresolvable path component") {
			t.Fatalf("expected unresolvable path error, got: %v", err)
		}
	})

	t.Run("regular file in path rejected", func(t *testing.T) {
		_, _, err := resolveSosDestinationDir(baseDir, "regular-file")
		if err == nil || !strings.Contains(err.Error(), "failed to create destination directory") {
			t.Fatalf("expected mkdir error, got: %v", err)
		}
	})

	t.Run("internal symlink within baseDir allowed", func(t *testing.T) {
		got, _, err := resolveSosDestinationDir(baseDir, "internal-link")
		if err != nil {
			t.Fatalf("unexpected error for internal symlink: %v", err)
		}
		want := filepath.Join(baseDir, "internal-link")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

type recordedCmd struct {
	Name string
	Args []string
}

func TestSosHelperProcess(_ *testing.T) {
	if os.Getenv("GO_WANT_SOS_HELPER_PROCESS") != "1" {
		return
	}
	scenario := os.Getenv("GO_SOS_SCENARIO")
	cmdName := os.Getenv("GO_SOS_CMD_NAME")
	var cmdArgs []string
	if err := json.Unmarshal([]byte(os.Getenv("GO_SOS_CMD_ARGS")), &cmdArgs); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid helper args: %v\n", err)
		os.Exit(2)
	}

	if scenario == "all-fail" {
		_, _ = fmt.Fprintln(os.Stderr, "simulated failure")
		os.Exit(1)
	}

	switch cmdName {
	case "kubectl":
		if len(cmdArgs) > 0 && cmdArgs[0] == "get" {
			if strings.HasPrefix(scenario, "pod") {
				_, _ = fmt.Fprint(os.Stdout, "True")
			}
			os.Exit(0)
		}
		if len(cmdArgs) >= 3 && cmdArgs[0] == "exec" {
			if slices.Contains(cmdArgs, "cat") {
				if scenario == "pod-cat-fail" {
					_, _ = fmt.Fprint(os.Stderr, "simulated cat failure")
					os.Exit(1)
				}
				_, _ = fmt.Fprint(os.Stdout, "fake-pod-sos-archive")
				os.Exit(0)
			}
			for _, a := range cmdArgs {
				if strings.Contains(a, "sos report") {
					// Extract /tmp/sos-sos-debug-... from --tmp-dir=/host/tmp/sos-sos-debug-...
					idx := strings.Index(a, "/tmp/sos-")
					if idx >= 0 {
						tmpDir := strings.Fields(a[idx:])[0]
						switch scenario {
						case "pod-bad-remote":
							_, _ = fmt.Fprintf(os.Stdout, "Your sosreport has been generated and saved in: %s/;id;.tar.xz\n", tmpDir)
						case "pod-dotdot-only":
							_, _ = fmt.Fprintf(os.Stdout, "Your sosreport has been generated and saved in: %s/sosreport-a..b.tar.xz\n", tmpDir)
						case "pod-dotdot-then-valid":
							_, _ = fmt.Fprintf(os.Stdout, "ignore %s/sosreport-a..b.tar.xz\nsaved in: %s/sosreport-gke-node-1-2026-01-01-abc.tar.xz\n", tmpDir, tmpDir)
						default:
							_, _ = fmt.Fprintf(os.Stdout, "Your sosreport has been generated and saved in: %s/sosreport-gke-node-1-2026-01-01-abc.tar.xz\n", tmpDir)
						}
					}
					os.Exit(0)
				}
			}
			os.Exit(0)
		}
		os.Exit(0)
	case "gcloud":
		if len(cmdArgs) >= 2 && cmdArgs[0] == "compute" {
			switch cmdArgs[1] {
			case "instances":
				switch scenario {
				case "ssh-bad-zone-flag":
					_, _ = fmt.Fprintln(os.Stdout, "--project=evil")
				case "ssh-bad-zone-multiline":
					_, _ = fmt.Fprintln(os.Stdout, "us-central1-a\nus-central1-b")
				default:
					_, _ = fmt.Fprintln(os.Stdout, "us-central1-a")
				}
				os.Exit(0)
			case "ssh":
				if slices.Contains(cmdArgs, "sudo sos report --all-logs --batch --tmp-dir=/var") {
					switch scenario {
					case "ssh-bad-remote":
						_, _ = fmt.Fprintln(os.Stdout, "saved in /var/sosreport-;id;.tar.xz")
					case "ssh-dotdot-only":
						_, _ = fmt.Fprintln(os.Stdout, "saved in /var/sosreport-a..b.tar.xz")
					case "ssh-dotdot-then-valid":
						_, _ = fmt.Fprintln(os.Stdout, "ignore /var/sosreport-a..b.tar.xz\nsaved in /var/sosreport-gke-node-1-2026-01-01-abc.tar.xz")
					default:
						_, _ = fmt.Fprintln(os.Stdout, "saved in /var/sosreport-gke-node-1-2026-01-01-abc.tar.xz")
					}
				}
				os.Exit(0)
			case "scp":
				dst := cmdArgs[len(cmdArgs)-1]
				if !filepath.IsAbs(dst) {
					os.Exit(2)
				}
				// #nosec G304,G306,G703 -- test helper writes to the validated temp path passed to scp.
				if err := os.WriteFile(dst, []byte("fake-ssh-sos-archive"), 0o644); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "helper write failed: %v\n", err)
					os.Exit(2)
				}
				// #nosec G302,G703 -- test helper intentionally sets 0644 to verify SSH handler tightens mode to 0600.
				if err := os.Chmod(dst, 0o644); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "helper chmod failed: %v\n", err)
					os.Exit(2)
				}
				os.Exit(0)
			}
		}
	}
	os.Exit(1)
}

func stubClusterExecCommand(t *testing.T, scenario string, recorded *[]recordedCmd) {
	t.Helper()
	orig := execCommand
	t.Cleanup(func() {
		execCommand = orig
	})
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		*recorded = append(*recorded, recordedCmd{
			Name: name,
			Args: append([]string(nil), args...),
		})
		argsJSON, _ := json.Marshal(args)
		cmdArgs := []string{"-test.run=TestSosHelperProcess", "--"}
		// #nosec G204,G702 -- re-execs the test binary itself with fixed helper arguments.
		cmd := exec.CommandContext(ctx, os.Args[0], cmdArgs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_SOS_HELPER_PROCESS=1",
			"GO_SOS_SCENARIO="+scenario,
			"GO_SOS_CMD_NAME="+name,
			"GO_SOS_CMD_ARGS="+string(argsJSON),
		)
		return cmd
	}
}

func TestGetNodeSosReport(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() failed: %v", err)
	}

	h := &handlers{}

	t.Run("malicious destinations rejected without running commands", func(t *testing.T) {
		for _, badDest := range []string{
			"/etc/x",
			"--foo",
			"-x",
			"../outside",
			"./-foo",
			"   ",
			"bad\x00dir",
			"bad\ndir",
		} {
			var cmds []recordedCmd
			stubClusterExecCommand(t, "pod-ok", &cmds)

			_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
				Node:        "gke-node-1",
				Destination: badDest,
			})
			if err == nil {
				t.Errorf("expected error for destination %q, got nil", badDest)
			}
			if len(cmds) != 0 {
				t.Errorf("expected no commands for %q, got %v", badDest, cmds)
			}
		}
	})

	t.Run("invalid node names rejected without running commands", func(t *testing.T) {
		for _, badNode := range []string{"", "-bad-node", "bad;node", "a"} {
			var cmds []recordedCmd
			stubClusterExecCommand(t, "pod-ok", &cmds)

			_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
				Node: badNode,
			})
			if err == nil {
				t.Errorf("expected error for node %q, got nil", badNode)
			}
			if len(cmds) != 0 {
				t.Errorf("expected no commands for node %q, got %v", badNode, cmds)
			}
		}
	})

	t.Run("omitted destination cleans up temp dir on failure", func(t *testing.T) {
		tmpRoot := t.TempDir()
		t.Setenv("TMPDIR", tmpRoot)

		var cmds []recordedCmd
		stubClusterExecCommand(t, "all-fail", &cmds)

		_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node: "gke-node-1",
		})
		if err == nil {
			t.Fatal("expected error when commands fail, got nil")
		}
		entries, readErr := os.ReadDir(tmpRoot)
		if readErr != nil {
			t.Fatalf("ReadDir(%q) failed: %v", tmpRoot, readErr)
		}
		if len(entries) != 0 {
			t.Errorf("expected temp dir to be cleaned up on failure, found %d entries", len(entries))
		}
	})

	t.Run("omitted destination succeeds and preserves temp report", func(t *testing.T) {
		tmpRoot := t.TempDir()
		t.Setenv("TMPDIR", tmpRoot)

		var cmds []recordedCmd
		stubClusterExecCommand(t, "pod-ok", &cmds)

		res, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:   "gke-node-1",
			Method: "pod",
		})
		if err != nil {
			t.Fatalf("getNodeSosReport() unexpected error: %v", err)
		}
		text, ok := res.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("unexpected content type %T", res.Content[0])
		}
		const prefix = "SOS report successfully generated and downloaded to: "
		if !strings.HasPrefix(text.Text, prefix) {
			t.Fatalf("unexpected response %q", text.Text)
		}
		savedPath := strings.TrimPrefix(text.Text, prefix)
		rel, relErr := filepath.Rel(tmpRoot, savedPath)
		if relErr != nil || !filepath.IsLocal(rel) {
			t.Errorf("savedPath %q is not inside TMPDIR %q", savedPath, tmpRoot)
		}
		// #nosec G304 -- savedPath was verified above to reside within t.TempDir() TMPDIR.
		data, readErr := os.ReadFile(savedPath)
		if readErr != nil {
			t.Fatalf("expected downloaded report at %q to exist, err = %v", savedPath, readErr)
		}
		if string(data) != "fake-pod-sos-archive" {
			t.Errorf("unexpected file contents %q", data)
		}
	})

	t.Run("explicit nested destination cleaned up on failure while existing dirs preserved", func(t *testing.T) {
		var cmds []recordedCmd
		stubClusterExecCommand(t, "all-fail", &cmds)

		_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: filepath.Join("left", "behind", "tree"),
		})
		if err == nil {
			t.Fatal("expected error when commands fail, got nil")
		}
		if _, statErr := os.Lstat(filepath.Join(cwd, "left")); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("expected newly created tree under %q to be removed, statErr = %v", filepath.Join(cwd, "left"), statErr)
		}

		existing := filepath.Join(cwd, "already-exists")
		if mkErr := os.Mkdir(existing, 0o750); mkErr != nil {
			t.Fatalf("Mkdir failed: %v", mkErr)
		}
		keepFile := filepath.Join(existing, "keep.txt")
		if wfErr := os.WriteFile(keepFile, []byte("keep"), 0o600); wfErr != nil {
			t.Fatalf("WriteFile failed: %v", wfErr)
		}
		_, _, err = h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "already-exists",
		})
		if err == nil {
			t.Fatal("expected error when commands fail, got nil")
		}
		if _, statErr := os.Stat(keepFile); statErr != nil {
			t.Errorf("expected pre-existing directory file %q to be preserved, statErr = %v", keepFile, statErr)
		}
	})

	t.Run("pod path downloads report to validated directory with restrictive permissions and unique names", func(t *testing.T) {
		var cmds []recordedCmd
		stubClusterExecCommand(t, "pod-ok", &cmds)

		res1, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "sos-pod-out",
			Method:      "pod",
		})
		if err != nil {
			t.Fatalf("first getNodeSosReport() unexpected error: %v", err)
		}
		res2, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "sos-pod-out",
			Method:      "pod",
		})
		if err != nil {
			t.Fatalf("second getNodeSosReport() unexpected error: %v", err)
		}
		if len(res1.Content) != 1 || len(res2.Content) != 1 {
			t.Fatalf("unexpected content counts: %d, %d", len(res1.Content), len(res2.Content))
		}
		text, ok := res1.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatalf("unexpected content type %T", res1.Content[0])
		}
		wantDir := filepath.Join(cwd, "sos-pod-out")
		if !strings.Contains(text.Text, wantDir) {
			t.Errorf("response %q missing destination dir %q", text.Text, wantDir)
		}

		entries, err := os.ReadDir(wantDir)
		if err != nil || len(entries) != 2 {
			t.Fatalf("expected 2 distinct files in %q, got %d (err=%v)", wantDir, len(entries), err)
		}
		info, err := entries[0].Info()
		if err != nil {
			t.Fatalf("Info() failed: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("file permissions = %o, want 0600", perm)
		}
		var sawCat bool
		for _, c := range cmds {
			if c.Name == "kubectl" && slices.Contains(c.Args, "cat") {
				sawCat = true
				catIdx := slices.Index(c.Args, "cat")
				if catIdx+1 >= len(c.Args) || c.Args[catIdx+1] != "--" {
					t.Errorf("kubectl cat missing -- delimiter: %v", c.Args)
				}
			}
		}
		if !sawCat {
			t.Errorf("expected kubectl exec cat invocation, got %v", cmds)
		}
	})

	t.Run("pod path cleans up partial file when cat fails", func(t *testing.T) {
		var cmds []recordedCmd
		stubClusterExecCommand(t, "pod-cat-fail", &cmds)

		_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "sos-pod-cat-fail",
			Method:      "pod",
		})
		if err == nil || !strings.Contains(err.Error(), "failed to copy sos report from pod") {
			t.Fatalf("expected cat failure, got: %v", err)
		}
		if _, statErr := os.Lstat(filepath.Join(cwd, "sos-pod-cat-fail")); !errors.Is(statErr, os.ErrNotExist) {
			t.Errorf("expected newly created destination to be cleaned up on cat failure, statErr = %v", statErr)
		}
	})

	t.Run("pod path rejects shell metacharacters and dotdot in remote filename and skips earlier dotdot", func(t *testing.T) {
		for _, scenario := range []string{"pod-bad-remote", "pod-dotdot-only"} {
			var cmds []recordedCmd
			stubClusterExecCommand(t, scenario, &cmds)

			_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
				Node:        "gke-node-1",
				Destination: "sos-pod-bad",
				Method:      "pod",
			})
			if err == nil || !strings.Contains(err.Error(), "could not find sos report filename") {
				t.Fatalf("scenario %s: expected filename parse error, got: %v", scenario, err)
			}
			for _, c := range cmds {
				if slices.Contains(c.Args, "cat") {
					t.Errorf("scenario %s: kubectl cat should not run on invalid remote filename: %v", scenario, c.Args)
				}
			}
		}

		var cmds []recordedCmd
		stubClusterExecCommand(t, "pod-dotdot-then-valid", &cmds)
		if _, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "sos-pod-skip-dotdot",
			Method:      "pod",
		}); err != nil {
			t.Fatalf("pod-dotdot-then-valid unexpected error: %v", err)
		}
	})

	t.Run("ssh path passes end-of-options separator and enforces 0600 permissions", func(t *testing.T) {
		var cmds []recordedCmd
		stubClusterExecCommand(t, "ssh-ok", &cmds)

		res, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "sos-ssh-out",
			Method:      "ssh",
		})
		if err != nil {
			t.Fatalf("getNodeSosReport() unexpected error: %v", err)
		}
		wantDir := filepath.Join(cwd, "sos-ssh-out")
		text, ok := res.Content[0].(*mcp.TextContent)
		if !ok || !strings.Contains(text.Text, wantDir) {
			t.Errorf("response %#v missing destination dir %q", res.Content[0], wantDir)
		}

		entries, err := os.ReadDir(wantDir)
		if err != nil || len(entries) != 1 {
			t.Fatalf("expected 1 file in %q, got %d (err=%v)", wantDir, len(entries), err)
		}
		info, err := entries[0].Info()
		if err != nil {
			t.Fatalf("Info() failed: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Errorf("ssh downloaded file permissions = %o, want 0600", perm)
		}

		var sawSCP bool
		for _, c := range cmds {
			if c.Name == "gcloud" && len(c.Args) >= 2 && c.Args[1] == "scp" {
				sawSCP = true
				sepIdx := slices.Index(c.Args, "--")
				if sepIdx == -1 || sepIdx+2 >= len(c.Args) {
					t.Fatalf("gcloud scp missing -- before positional args: %v", c.Args)
				}
				src := c.Args[sepIdx+1]
				dst := c.Args[sepIdx+2]
				if src != "gke-node-1:/var/sosreport-gke-node-1-2026-01-01-abc.tar.xz" {
					t.Errorf("scp src = %q, unexpected", src)
				}
				if !strings.HasPrefix(dst, wantDir+string(filepath.Separator)) {
					t.Errorf("scp dst = %q, want under %q", dst, wantDir)
				}
			}
		}
		if !sawSCP {
			t.Errorf("expected gcloud compute scp invocation, got %v", cmds)
		}
	})

	t.Run("ssh path rejects shell metacharacters and dotdot in remote filename and skips earlier dotdot", func(t *testing.T) {
		for _, scenario := range []string{"ssh-bad-remote", "ssh-dotdot-only"} {
			var cmds []recordedCmd
			stubClusterExecCommand(t, scenario, &cmds)

			_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
				Node:        "gke-node-1",
				Destination: "sos-ssh-bad-remote",
				Method:      "ssh",
			})
			if err == nil || !strings.Contains(err.Error(), "could not find sos report filename") {
				t.Fatalf("scenario %s: expected filename parse error, got: %v", scenario, err)
			}
			for _, c := range cmds {
				if c.Name == "gcloud" && len(c.Args) >= 2 && c.Args[1] == "scp" {
					t.Errorf("scenario %s: gcloud scp should not be called on invalid remote filename: %v", scenario, c.Args)
				}
			}
		}

		var cmds []recordedCmd
		stubClusterExecCommand(t, "ssh-dotdot-then-valid", &cmds)
		if _, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
			Node:        "gke-node-1",
			Destination: "sos-ssh-skip-dotdot",
			Method:      "ssh",
		}); err != nil {
			t.Fatalf("ssh-dotdot-then-valid unexpected error: %v", err)
		}
	})

	t.Run("openSosOutputFile rejects pre-existing file with O_EXCL", func(t *testing.T) {
		target := filepath.Join(cwd, "already-present.tar.xz")
		if err := os.WriteFile(target, []byte("existing"), 0o600); err != nil {
			t.Fatalf("WriteFile failed: %v", err)
		}
		f, err := openSosOutputFile(target)
		if err == nil {
			_ = f.Close()
			t.Fatal("expected error opening pre-existing file with O_EXCL, got nil")
		}
		if !errors.Is(err, os.ErrExist) {
			t.Errorf("expected os.ErrExist, got %v", err)
		}
	})

	t.Run("ssh path rejects invalid zone output before ssh", func(t *testing.T) {
		for _, scenario := range []string{"ssh-bad-zone-flag", "ssh-bad-zone-multiline"} {
			var cmds []recordedCmd
			stubClusterExecCommand(t, scenario, &cmds)

			_, _, err := h.getNodeSosReport(context.Background(), &mcp.CallToolRequest{}, &getNodeSosReportArgs{
				Node:        "gke-node-1",
				Destination: "sos-ssh-bad-zone",
				Method:      "ssh",
			})
			if err == nil || !strings.Contains(err.Error(), "invalid zone") {
				t.Fatalf("scenario %s: expected invalid zone error, got: %v", scenario, err)
			}
			for _, c := range cmds {
				if c.Name == "gcloud" && len(c.Args) >= 2 && c.Args[1] == "ssh" {
					t.Errorf("scenario %s: gcloud ssh should not be called on invalid zone: %v", scenario, c.Args)
				}
			}
		}
	})
}

func TestGetNodeSosReport_MCPDispatch(t *testing.T) {
	workDir := t.TempDir()
	t.Chdir(workDir)

	h := &handlers{}
	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "1.0.0"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_node_sos_report",
		Description: "test tool",
	}, h.getNodeSosReport)

	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = server.Run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() failed: %v", err)
	}
	defer func() { _ = session.Close() }()

	toolsRes, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil || len(toolsRes.Tools) != 1 {
		t.Fatalf("ListTools() failed: tools=%v err=%v", toolsRes, err)
	}
	schemaBytes, err := json.Marshal(toolsRes.Tools[0].InputSchema)
	if err != nil {
		t.Fatalf("json.Marshal(InputSchema) failed: %v", err)
	}
	if !strings.Contains(string(schemaBytes), "MCP server's current working directory") {
		t.Errorf("InputSchema %s does not specify MCP server's working directory", schemaBytes)
	}

	for _, badDest := range []string{"/etc/x", "--foo"} {
		var cmds []recordedCmd
		stubClusterExecCommand(t, "ssh-ok", &cmds)

		res, err := session.CallTool(ctx, &mcp.CallToolParams{
			Name: "get_node_sos_report",
			Arguments: map[string]any{
				"node":        "gke-node-1",
				"destination": badDest,
			},
		})
		if err != nil {
			t.Fatalf("CallTool(%q) protocol error: %v", badDest, err)
		}
		if !res.IsError {
			t.Errorf("CallTool(%q) IsError = false, want true", badDest)
		}
		if len(cmds) != 0 {
			t.Errorf("commands were invoked for %q: %v", badDest, cmds)
		}
	}
}
