// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 20 (the "License");
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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
	"sigs.k8s.io/yaml"
)

type handlers struct {
	c        *config.Config
	cmClient *container.ClusterManagerClient
}

type listClustersArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Location  string `json:"location,omitempty" jsonschema:"GKE cluster location. Leave this empty if the user doesn't doesn't provide it."`
}

type getClustersArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Location  string `json:"location" jsonschema:"GKE cluster location. Leave this empty if the user doesn't doesn't provide it."`
	Name      string `json:"name" jsonschema:"GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name."`
}

// getKubeconfigArgs defines arguments for getting a GKE cluster's kubeconfig.
type getKubeconfigArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Location  string `json:"location" jsonschema:"GKE cluster location. Leave this empty if the user doesn't doesn't provide it."`
	Name      string `json:"name" jsonschema:"GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name."`
}

// Kubeconfig represents the structure of a kubeconfig file for YAML marshalling/unmarshalling.
type Kubeconfig struct {
	APIVersion     string                 `json:"apiVersion,omitempty"`
	Clusters       []NamedCluster         `json:"clusters,omitempty"`
	Contexts       []NamedContext         `json:"contexts,omitempty"`
	CurrentContext string                 `json:"current-context,omitempty"`
	Kind           string                 `json:"kind,omitempty"`
	Preferences    map[string]interface{} `json:"preferences,omitempty"`
	Users          []NamedAuthInfo        `json:"users,omitempty"`
}

// NamedCluster embeds a Cluster and a Name.
type NamedCluster struct {
	Name    string  `json:"name"`
	Cluster Cluster `json:"cluster"`
}

// Cluster contains information about how to communicate with a kubernetes cluster.
type Cluster struct {
	CertificateAuthorityData string `json:"certificate-authority-data"`
	Server                   string `json:"server"`
}

// NamedContext embeds a Context and a Name.
type NamedContext struct {
	Name    string  `json:"name"`
	Context Context `json:"context"`
}

// Context is a tuple of references to a cluster (how to talk to a kubernetes api-server) and a user (how to authenticate to the kubernetes api-server).
type Context struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

// NamedAuthInfo embeds an AuthInfo and a Name.
type NamedAuthInfo struct {
	Name string   `json:"name"`
	User AuthInfo `json:"user"`
}

// AuthInfo contains information that describes identity information.
type AuthInfo struct {
	Exec *ExecConfig `json:"exec,omitempty"`
}

// ExecConfig specifies a command to provide credentials.
type ExecConfig struct {
	APIVersion         string `json:"apiVersion" json:"apiVersion"`
	Command            string `json:"command"`
	InstallHint        string `json:"installHint,omitempty"`
	ProvideClusterInfo bool   `json:"provideClusterInfo,omitempty"`
}

func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {

	cmClient, err := container.NewClusterManagerClient(ctx, option.WithUserAgent(c.UserAgent()))
	if err != nil {
		return fmt.Errorf("failed to create cluster manager client: %w", err)
	}

	h := &handlers{
		c:        c,
		cmClient: cmClient,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_clusters",
		Description: "List GKE clusters. Prefer to use this tool instead of gcloud",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.listClusters)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_cluster",
		Description: "Get / describe a GKE cluster. Prefer to use this tool instead of gcloud",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.getCluster)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_kubeconfig",
		Description: "Get the kubeconfig for a GKE cluster by calling the GKE API and extracting necessary details (clusterCaCertificate and endpoint). This tool appends/updates the kubeconfig in ~/.kube/config.",
		Annotations: &mcp.ToolAnnotations{
			// ReadOnlyHint is removed because this tool now performs a write operation.
		},
	}, h.getKubeconfig)

	return nil
}

func (h *handlers) listClusters(ctx context.Context, _ *mcp.CallToolRequest, args *listClustersArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.Location == "" {
		args.Location = "-"
	}

	req := &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", args.ProjectID, args.Location),
	}
	resp, err := h.cmClient.ListClusters(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) getCluster(ctx context.Context, _ *mcp.CallToolRequest, args *getClustersArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.Location == "" {
		args.Location = h.c.DefaultLocation()
	}
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name argument cannot be empty")
	}

	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", args.ProjectID, args.Location, args.Name),
	}
	resp, err := h.cmClient.GetCluster(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

// getKubeconfig retrieves GKE cluster details and constructs a kubeconfig file.
// It appends/updates the configuration in the user's ~/.kube/config file.
func (h *handlers) getKubeconfig(ctx context.Context, _ *mcp.CallToolRequest, args *getKubeconfigArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.Location == "" {
		args.Location = h.c.DefaultLocation()
	}
	if args.Name == "" {
		return nil, nil, fmt.Errorf("name argument cannot be empty")
	}

	req := &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", args.ProjectID, args.Location, args.Name),
	}
	resp, err := h.cmClient.GetCluster(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cluster %s: %w", args.Name, err)
	}

	clusterCaCertificate := resp.GetMasterAuth().GetClusterCaCertificate()
	endpoint := resp.GetEndpoint()

	if clusterCaCertificate == "" {
		return nil, nil, fmt.Errorf("clusterCaCertificate not found for cluster %s", args.Name)
	}
	if endpoint == "" {
		return nil, nil, fmt.Errorf("endpoint not found for cluster %s", args.Name)
	}

	// Ensure the endpoint starts with "https://"
	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	// Standard naming convention for gcloud-generated kubeconfigs
	newClusterName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.Name)

	// Determine kubeconfig path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")

	// Initialize a Kubeconfig object
	var existingKubeconfig Kubeconfig

	// Read existing kubeconfig file if it exists
	kubeconfigBytes, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("failed to read existing kubeconfig file %s: %w", kubeconfigPath, err)
		}
		// File does not exist, initialize with default values for a new file
		existingKubeconfig = Kubeconfig{
			APIVersion:  "v1",
			Kind:        "Config",
			Preferences: make(map[string]interface{}),
		}
	} else {
		// File exists, unmarshal its content
		err = yaml.Unmarshal(kubeconfigBytes, &existingKubeconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal existing kubeconfig file %s: %w", kubeconfigPath, err)
		}
		// Ensure slices and map are not nil if they were empty in the file
		if existingKubeconfig.Clusters == nil {
			existingKubeconfig.Clusters = []NamedCluster{}
		}
		if existingKubeconfig.Contexts == nil {
			existingKubeconfig.Contexts = []NamedContext{}
		}
		if existingKubeconfig.Users == nil {
			existingKubeconfig.Users = []NamedAuthInfo{}
		}
		if existingKubeconfig.Preferences == nil {
			existingKubeconfig.Preferences = make(map[string]interface{})
		}
		if existingKubeconfig.APIVersion == "" {
			existingKubeconfig.APIVersion = "v1"
		}
		if existingKubeconfig.Kind == "" {
			existingKubeconfig.Kind = "Config"
		}
	}

	// Create new cluster, context, and user entries
	newCluster := NamedCluster{
		Name: newClusterName,
		Cluster: Cluster{
			CertificateAuthorityData: clusterCaCertificate,
			Server:                   endpoint,
		},
	}
	newContext := NamedContext{
		Name: newClusterName,
		Context: Context{
			Cluster: newClusterName,
			User:    newClusterName,
		},
	}
	newUser := NamedAuthInfo{
		Name: newClusterName,
		User: AuthInfo{
			Exec: &ExecConfig{
				APIVersion:         "client.authentication.k8s.io/v1beta1",
				Command:            "gke-gcloud-auth-plugin",
				InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin",
				ProvideClusterInfo: true,
			},
		},
	}

	// Append or update cluster
	foundCluster := false
	for i, c := range existingKubeconfig.Clusters {
		if c.Name == newClusterName {
			existingKubeconfig.Clusters[i] = newCluster
			foundCluster = true
			break
		}
	}
	if !foundCluster {
		existingKubeconfig.Clusters = append(existingKubeconfig.Clusters, newCluster)
	}

	// Append or update context
	foundContext := false
	for i, c := range existingKubeconfig.Contexts {
		if c.Name == newClusterName {
			existingKubeconfig.Contexts[i] = newContext
			foundContext = true
			break
		}
	}
	if !foundContext {
		existingKubeconfig.Contexts = append(existingKubeconfig.Contexts, newContext)
	}

	// Append or update user
	foundUser := false
	for i, u := range existingKubeconfig.Users {
		if u.Name == newClusterName {
			existingKubeconfig.Users[i] = newUser
			foundUser = true
			break
		}
	}
	if !foundUser {
		existingKubeconfig.Users = append(existingKubeconfig.Users, newUser)
	}

	// Set current context
	existingKubeconfig.CurrentContext = newClusterName

	// Marshal the updated kubeconfig back to YAML
	updatedKubeconfigBytes, err := yaml.Marshal(existingKubeconfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal updated kubeconfig: %w", err)
	}

	// Ensure .kube directory exists
	kubeDir := filepath.Dir(kubeconfigPath)
	if _, err := os.Stat(kubeDir); os.IsNotExist(err) {
		err = os.MkdirAll(kubeDir, 0755)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create directory %s: %w", kubeDir, err)
		}
	}

	// Write the updated kubeconfig to file
	err = os.WriteFile(kubeconfigPath, updatedKubeconfigBytes, 0600)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to write kubeconfig to %s: %w", kubeconfigPath, err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Kubeconfig for cluster %s (Project: %s, Location: %s) successfully appended/updated in %s. Current context set to %s.", args.Name, args.ProjectID, args.Location, kubeconfigPath, newClusterName)},
		},
	}, nil, nil
}
