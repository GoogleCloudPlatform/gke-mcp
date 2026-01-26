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

// Package cluster provides MCP tools for managing GKE clusters.
package cluster

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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

type createClustersArgs struct {
	ProjectID string              `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Location  string              `json:"location" jsonschema:"GKE cluster location. Leave this empty if the user doesn't doesn't provide it."`
	Cluster   containerpb.Cluster `json:"cluster" jsonschema:"GKE cluster configuration."`
}

// getKubeconfigArgs defines arguments for getting a GKE cluster's kubeconfig.
type getKubeconfigArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Location  string `json:"location" jsonschema:"GKE cluster location. Leave this empty if the user doesn't provide it."`
	Name      string `json:"name" jsonschema:"GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name."`
}

type getNodeSosReportArgs struct {
	Node           string `json:"node" jsonschema:"GKE node name to collect SOS report from."`
	Destination    string `json:"destination,omitempty" jsonschema:"Local directory to download the SOS report to. Defaults to /tmp/sos-report if not specified."`
	Method         string `json:"method,omitempty" jsonschema:"Method to get sos report. Can be 'pod', 'ssh' or 'any'. Defaults to 'any'. When the node is unhealthy from api server, use ssh only."`
	TimeoutSeconds int    `json:"timeout,omitempty" jsonschema:"Timeout in seconds for the report collection (applies to both pod and ssh methods). Defaults to 180 (3 minutes)."`
}

// Install registers cluster-related tools with the MCP server.
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
		Name:        "create_cluster",
		Description: "Create a GKE cluster. Prefer to use this tool instead of gcloud",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: false,
		},
	}, h.createCluster)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_kubeconfig",
		Description: "Get the kubeconfig for a GKE cluster by calling the GKE API and extracting necessary details (clusterCaCertificate and endpoint). This tool appends/updates the kubeconfig in ~/.kube/config.",
		Annotations: &mcp.ToolAnnotations{
			// ReadOnlyHint is removed because this tool now performs a write operation.
		},
	}, h.getKubeconfig)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_node_sos_report",
		Description: "Generate and download an SOS report from a GKE node. Can use 'pod', 'ssh' or 'any' methods. Defaults to 'any' (pod with fallback to ssh). Use 'ssh' if node is API-unhealthy.",
	}, h.getNodeSosReport)

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

	header := fmt.Sprintf("Found %d clusters in project %s:", len(resp.Clusters), args.ProjectID)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: header},
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

func (h *handlers) createCluster(ctx context.Context, _ *mcp.CallToolRequest, args *createClustersArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.Location == "" {
		args.Location = h.c.DefaultLocation()
	}

	req := &containerpb.CreateClusterRequest{
		Parent:  fmt.Sprintf("projects/%s/locations/%s", args.ProjectID, args.Location),
		Cluster: &args.Cluster,
	}
	resp, err := h.cmClient.CreateCluster(ctx, req)
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

func (h *handlers) getNodeSosReport(ctx context.Context, _ *mcp.CallToolRequest, args *getNodeSosReportArgs) (*mcp.CallToolResult, any, error) {
	if args.Node == "" {
		return nil, nil, fmt.Errorf("node argument cannot be empty")
	}
	// Basic validation for node name to prevent command injection
	if match, _ := regexp.MatchString(`^[a-z0-9][a-z0-9\-\.]*[a-z0-9]$`, args.Node); !match {
		return nil, nil, fmt.Errorf("invalid node name: %s", args.Node)
	}

	if args.Destination == "" {
		args.Destination = "/tmp/sos-report"
	}
	if args.Method == "" {
		args.Method = "any"
	}
	if args.TimeoutSeconds <= 0 {
		args.TimeoutSeconds = 180 // Default to 3 minutes
	}

	// Check if node is healthy
	isHealthy := false
	// #nosec G204
	cmd := exec.CommandContext(ctx, "kubectl", "get", "node", args.Node, "-o", "jsonpath='{.status.conditions[?(@.type==\"Ready\")].status}'")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "True") {
		isHealthy = true
	}

	if !isHealthy {
		args.Method = "ssh"
	}

	if err := os.MkdirAll(args.Destination, 0750); err != nil {
		return nil, nil, fmt.Errorf("failed to create destination directory %s: %w", args.Destination, err)
	}

	if args.Method == "pod" || args.Method == "any" {
		// 1. Try Pod-based approach with timeout
		podCtx, podCancel := context.WithTimeout(ctx, time.Duration(args.TimeoutSeconds)*time.Second)
		defer podCancel()

		res, _, err := h.getNodeSosReportWithPod(podCtx, args)
		if err == nil {
			return res, nil, nil
		}
		if args.Method == "pod" {
			return nil, nil, fmt.Errorf("failed to get sos report with pod: %w", err)
		}
		// If method is any and pod failed (e.g. timeout), fall through to ssh
	}

	// 2. Fallback or direct SSH approach with timeout
	sshCtx, sshCancel := context.WithTimeout(ctx, time.Duration(args.TimeoutSeconds)*time.Second)
	defer sshCancel()
	return h.getNodeSosReportWithSSH(sshCtx, args)
}

func (h *handlers) getNodeSosReportWithPod(ctx context.Context, args *getNodeSosReportArgs) (*mcp.CallToolResult, any, error) {
	// 1. Prepare and run debug pod
	podName := fmt.Sprintf("sos-debug-%d", time.Now().Unix())
	overrides := map[string]interface{}{
		"spec": map[string]interface{}{
			"nodeName":    args.Node,
			"hostNetwork": true,
			"hostPID":     true,
			"hostIPC":     true,
			"containers": []map[string]interface{}{
				{
					"name":    "main",
					"image":   "gke.gcr.io/debian-base",
					"command": []string{"/bin/sleep", "99999"},
					"volumeMounts": []map[string]interface{}{
						{
							"mountPath": "/host",
							"name":      "root",
						},
					},
				},
			},
			"volumes": []map[string]interface{}{
				{
					"name": "root",
					"hostPath": map[string]interface{}{
						"path": "/",
						"type": "Directory",
					},
				},
			},
			"securityContext": map[string]interface{}{
				"runAsUser": 0,
			},
			"nodeSelector": map[string]interface{}{
				"kubernetes.io/hostname": args.Node,
			},
		},
	}
	overridesBytes, err := json.Marshal(overrides)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal overrides: %w", err)
	}

	// #nosec G204
	runCmd := exec.CommandContext(ctx, "kubectl", "run", podName, "--image=gke.gcr.io/debian-base", "--restart=Never", "--overrides="+string(overridesBytes))
	if out, err := runCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to create debug pod: %s, %w", string(out), err)
	}

	defer func() {
		// Cleanup pod
		// #nosec G204
		delCmd := exec.Command("kubectl", "delete", "pod", podName, "--wait=false", "--grace-period=0", "--force")
		_ = delCmd.Run() // Best-effort cleanup
	}()

	// 2. Wait for pod to be ready
	// #nosec G204
	waitCmd := exec.CommandContext(ctx, "kubectl", "wait", "--for=condition=Ready", "pod/"+podName, "--timeout=60s")
	if out, err := waitCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("debug pod did not become ready: %s, %w", string(out), err)
	}

	// 3. Run sos report inside the pod (chrooted to host)
	// We create a temp dir for the report to avoid conflicts in /tmp
	remoteTmpDir := fmt.Sprintf("/tmp/sos-%s", podName)
	// Prepare command: mkdir dir, run sos report, and ensure we capture output
	// Note: chroot /host allows us to use the host's sosreport command and filesystem
	execScript := fmt.Sprintf("apt update && apt install -y sosreport && mkdir -p /host%s && sos report --sysroot=/host --all-logs --batch --tmp-dir=/host%s", remoteTmpDir, remoteTmpDir)

	// #nosec G204
	execCmd := exec.CommandContext(ctx, "kubectl", "exec", podName, "--", "sh", "-c", execScript)
	outBytes, err := execCmd.CombinedOutput()
	output := string(outBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate sos report: %s, %w", output, err)
	}

	// 4. Parse the output to find the filename
	// The output usually contains: "Your sosreport has been generated and saved in: /path/to/file.tar.xz"
	// The path might be reported as /host/tmp/... or /tmp/... depending on how sos report was invoked.
	// We also handle both .tar.xz and .tar.gz extensions.
	re := regexp.MustCompile(`(/host)?` + regexp.QuoteMeta(remoteTmpDir) + `/[^\s]+\.tar\.(xz|gz)`)
	match := re.FindString(output)
	if match == "" {
		return nil, nil, fmt.Errorf("could not find sos report filename in output: %s", output)
	}

	// The file path inside the pod is what we need for 'cat'.
	// If the match didn't start with /host, we prepend it because we mounted host root at /host.
	remotePath := match
	if !strings.HasPrefix(remotePath, "/host") {
		remotePath = "/host" + remotePath
	}
	localFilename := fmt.Sprintf("sosreport-%s-%s.tar.xz", args.Node, time.Now().Format("2006-01-02-15-04-05"))
	localPath := filepath.Join(args.Destination, localFilename)
	localPath = filepath.Clean(localPath)
	// 5. Copy the file from the pod to local current directory
	f, err := os.Create(localPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}

	// #nosec G204
	catCmd := exec.CommandContext(ctx, "kubectl", "exec", podName, "--", "cat", remotePath)
	catCmd.Stdout = f
	var stderr bytes.Buffer
	catCmd.Stderr = &stderr

	if err := catCmd.Run(); err != nil {
		_ = f.Close()            // Best-effort cleanup
		_ = os.Remove(localPath) // Best-effort cleanup
		return nil, nil, fmt.Errorf("failed to copy sos report from pod: %s, %w", stderr.String(), err)
	}
	_ = f.Close() // Best-effort cleanup

	// 6. Cleanup remote files on host (via pod)
	cleanupScript := fmt.Sprintf("rm -rf %s", remoteTmpDir)
	// #nosec G204
	cleanCmd := exec.CommandContext(ctx, "kubectl", "exec", podName, "--", "sh", "-c", cleanupScript)
	_ = cleanCmd.Run() // Best-effort cleanup

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("SOS report successfully generated and downloaded to: %s", localPath)},
		},
	}, nil, nil
}

func (h *handlers) getNodeSosReportWithSSH(ctx context.Context, args *getNodeSosReportArgs) (*mcp.CallToolResult, any, error) {
	// 1. Find the zone of the VM
	// gcloud compute instances list --filter="name=NODE_NAME" --format="value(zone)"
	// #nosec G204
	findZoneCmd := exec.CommandContext(ctx, "gcloud", "compute", "instances", "list", fmt.Sprintf("--filter=name=%s", args.Node), "--format=value(zone)")
	zoneOut, err := findZoneCmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find zone for node %s using gcloud: %w", args.Node, err)
	}
	zone := strings.TrimSpace(string(zoneOut))
	if zone == "" {
		return nil, nil, fmt.Errorf("could not find zone for node %s", args.Node)
	}

	// 2. Generate SOS report via SSH
	// gcloud compute ssh --zone "ZONE" "NODE_NAME" --command "sudo sos report --all-logs --batch --tmp-dir=/var"
	// #nosec G204
	sshCmd := exec.CommandContext(ctx, "gcloud", "compute", "ssh", "--zone", zone, args.Node, "--command", "sudo sos report --all-logs --batch --tmp-dir=/var")
	outBytes, err := sshCmd.CombinedOutput()
	output := string(outBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate sos report via ssh: %s, %w", output, err)
	}

	// 3. Parse output for filename
	// Matches /var/sosreport-.*.tar.xz
	re := regexp.MustCompile(`/var/sosreport-[^\s]+\.tar\.xz`)
	match := re.FindString(output)
	if match == "" {
		return nil, nil, fmt.Errorf("could not find sos report filename in ssh output: %s", output)
	}
	remotePath := match

	// 4. Change ownership of the file
	// gcloud compute ssh ... --command "sudo chown $USER REMOTE_PATH"
	// #nosec G204
	chownCmd := exec.CommandContext(ctx, "gcloud", "compute", "ssh", "--zone", zone, args.Node, "--command", fmt.Sprintf("sudo chown $USER %s", remotePath))
	if out, err := chownCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to chown remote file: %s, %w", string(out), err)
	}

	// 5. SCP the file
	// gcloud compute scp --zone "ZONE" "NODE_NAME:REMOTE_PATH" LOCAL_DESTINATION
	localFilename := fmt.Sprintf("sosreport-%s-%s.tar.xz", args.Node, time.Now().Format("2006-01-02-15-04-05"))
	localPath := filepath.Join(args.Destination, localFilename)
	// #nosec G204
	scpCmd := exec.CommandContext(ctx, "gcloud", "compute", "scp", "--zone", zone, fmt.Sprintf("%s:%s", args.Node, remotePath), localPath)
	if out, err := scpCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to scp file: %s, %w", string(out), err)
	}

	// 6. Cleanup remote files on host
	// #nosec G204
	rmCmd := exec.CommandContext(ctx, "gcloud", "compute", "ssh", "--zone", zone, args.Node, "--command", fmt.Sprintf("sudo rm %s", remotePath))
	_ = rmCmd.Run() // Best-effort cleanup

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("SOS report successfully generated (via SSH) and downloaded to: %s", localPath)},
		},
	}, nil, nil
}
