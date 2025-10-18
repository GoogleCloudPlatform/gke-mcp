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
	"encoding/base64"
	"fmt"
	"strings"

	container "cloud.google.com/go/container/apiv1"
	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/client-go/tools/clientcmd"
	k8sClientApi "k8s.io/client-go/tools/clientcmd/api"
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
	Location  string `json:"location" jsonschema:"GKE cluster location. Leave this empty if the user doesn't provide it."`
	Name      string `json:"name" jsonschema:"GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name."`
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

	// Initialize a Kubeconfig object
	pathOptions := clientcmd.NewDefaultPathOptions()
	oldKubeconfig, err := pathOptions.GetStartingConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get starting config: %w", err)
	}
	newKubeconfig := oldKubeconfig.DeepCopy()

	// Create new cluster, context, and user entries
	clusterCaCertificateByte, err := base64.RawStdEncoding.DecodeString(clusterCaCertificate)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode clusterCaCertificate: %w", err)
	}

	newCluster := &k8sClientApi.Cluster{
		CertificateAuthorityData: clusterCaCertificateByte,
		Server:                   endpoint,
	}
	newContext := &k8sClientApi.Context{
		Cluster:  newClusterName,
		AuthInfo: newClusterName,
	}
	newUser := &k8sClientApi.AuthInfo{
		Exec: &k8sClientApi.ExecConfig{
			APIVersion:         "client.authentication.k8s.io/v1beta1",
			Command:            "gke-gcloud-auth-plugin",
			InstallHint:        "Install gke-gcloud-auth-plugin for use with kubectl by following https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl#install_plugin",
			ProvideClusterInfo: true,
		},
	}

	// Append or update cluster, context, and user using map assignments
	newKubeconfig.Clusters[newClusterName] = newCluster
	newKubeconfig.Contexts[newClusterName] = newContext
	newKubeconfig.AuthInfos[newClusterName] = newUser

	// Set current context
	newKubeconfig.CurrentContext = newClusterName

	err = clientcmd.ModifyConfig(pathOptions, *newKubeconfig, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to modify kubeconfig: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Kubeconfig for cluster %s (Project: %s, Location: %s) successfully appended/updated in %s. Current context set to %s.", args.Name, args.ProjectID, args.Location, pathOptions.GlobalFile, newClusterName)},
		},
	}, nil, nil
}
