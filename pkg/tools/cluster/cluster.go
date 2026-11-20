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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	container "cloud.google.com/go/container/apiv1"
	containerpb "cloud.google.com/go/container/apiv1/containerpb"
	"github.com/googleapis/gax-go/v2/callctx"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	k8sClientApi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/k8s"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
)

var (
	execCommand    = exec.CommandContext
	validNodeRe    = regexp.MustCompile(`^[a-z0-9][a-z0-9\-\.]*[a-z0-9]$`)
	validZoneRe    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)
	sshSosReportRe = regexp.MustCompile(`/var/sosreport-[a-zA-Z0-9._-]+\.tar\.xz`)

	commonClustersFieldMasks = []string{
		// go/keep-sorted start
		"autopilot",
		"createTime",
		"currentMasterVersion",
		"currentNodeCount",
		"currentNodeVersion",
		"description",
		"endpoint",
		"fleet",
		"location",
		"name",
		"network",
		"nodePools.name",
		"releaseChannel",
		"resourceLabels",
		"selfLink",
		"status",
		"statusMessage",
		"subnetwork",
		// go/keep-sorted end
	}

	listClustersFieldMasks = append(prefixStrings(commonClustersFieldMasks, "clusters."),
		// go/keep-sorted start
		"missingZones",
		// go/keep-sorted end
	)

	getClusterFieldMasks = append(commonClustersFieldMasks,
		// go/keep-sorted start
		"nodePools.locations",
		"nodePools.status",
		"nodePools.version",
		// go/keep-sorted end
	)

	listClustersDefaultFieldMask = initializeDefaultFieldMask(listClustersFieldMasks)
	getClusterDefaultFieldMask   = initializeDefaultFieldMask(getClusterFieldMasks)
)

type handlers struct {
	c           *config.Config
	cmClient    *container.ClusterManagerClient
	k8sProvider k8s.Provider
}

type listClustersArgs struct {
	params.LocationOptional
	ReadMask string `json:"readMask,omitempty" jsonschema:"Optional. The field mask to specify the fields to be returned in the response. Use a single * to get all fields. Default: clusters.autopilot,clusters.createTime,clusters.currentMasterVersion,clusters.currentNodeCount,clusters.currentNodeVersion,clusters.description,clusters.endpoint,clusters.fleet,clusters.location,clusters.name,clusters.network,clusters.nodePools.name,clusters.releaseChannel,clusters.resourceLabels,clusters.selfLink,clusters.status,clusters.statusMessage,clusters.subnetwork,missingZones."`
}

type getClustersArgs struct {
	params.Cluster
	ReadMask string `json:"readMask,omitempty" jsonschema:"Optional. The field mask to specify the fields to be returned in the response. Use a single * to get all fields. Default: autopilot,createTime,currentMasterVersion,currentNodeCount,currentNodeVersion,description,endpoint,fleet,location,name,network,nodePools.locations,nodePools.name,nodePools.status,nodePools.version,releaseChannel,resourceLabels,selfLink,status,statusMessage,subnetwork."`
}

type createClustersArgs struct {
	params.LocationRequired
	Cluster string `json:"cluster" jsonschema:"Required. A cluster resource represented as a string using JSON format."`
}

type updateClusterArgs struct {
	params.Cluster
	Update string `json:"update" jsonschema:"Required. A description of the update represented as a string using JSON format."`
}

type deleteClusterArgs struct {
	params.Cluster
	DeletionPolicy string `json:"deletionPolicy,omitempty" jsonschema:"Optional. The deletion policy to apply to the request. Options: 'VERIFY_UNUSED' (default: checks for active compute, external exposure, and persistent data before deletion), 'FORCE' (immediately deletes the cluster without safety checks)."`
}

// getKubeconfigArgs defines arguments for getting a GKE cluster's kubeconfig.
type getKubeconfigArgs struct {
	params.Cluster
}

type getNodeSosReportArgs struct {
	params.Cluster
	Node           string `json:"node" jsonschema:"Required. GKE node name to collect SOS report from."`
	Destination    string `json:"destination,omitempty" jsonschema:"Relative path within the MCP server's current working directory to download the SOS report to. If omitted, downloads into a newly created temporary directory."`
	Method         string `json:"method,omitempty" jsonschema:"Method to get sos report. Can be 'pod', 'ssh' or 'any'. Defaults to 'any'. When the node is unhealthy from api server, use ssh only."`
	TimeoutSeconds int    `json:"timeout,omitempty" jsonschema:"Timeout in seconds for the report collection (applies to both pod and ssh methods). Defaults to 180 (3 minutes)."`
}

func (h *handlers) listClusters(ctx context.Context, _ *mcp.CallToolRequest, args *listClustersArgs) (*mcp.CallToolResult, any, error) {
	ctx = callctx.SetHeaders(ctx,
		callctx.XGoogFieldMaskHeader,
		getFieldMask(args.ReadMask, listClustersDefaultFieldMask))

	req := &containerpb.ListClustersRequest{
		Parent: args.LocationPath(),
	}
	resp, err := h.cmClient.ListClusters(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	header := fmt.Sprintf("Found %d clusters:", len(resp.Clusters))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: header},
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) getCluster(ctx context.Context, _ *mcp.CallToolRequest, args *getClustersArgs) (*mcp.CallToolResult, any, error) {
	ctx = callctx.SetHeaders(ctx,
		callctx.XGoogFieldMaskHeader,
		getFieldMask(args.ReadMask, getClusterDefaultFieldMask))

	req := &containerpb.GetClusterRequest{
		Name: args.ClusterPath(),
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
	var clusterObj containerpb.Cluster
	if err := protojson.Unmarshal([]byte(args.Cluster), &clusterObj); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal cluster JSON: %w", err)
	}

	req := &containerpb.CreateClusterRequest{
		Parent:  args.LocationPath(),
		Cluster: &clusterObj,
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
	req := &containerpb.GetClusterRequest{
		Name: args.ClusterPath(),
	}
	resp, err := h.cmClient.GetCluster(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cluster %s: %w", args.ClusterPath(), err)
	}

	clusterCaCertificate := resp.GetMasterAuth().GetClusterCaCertificate()
	endpoint := resp.GetEndpoint()

	if clusterCaCertificate == "" {
		return nil, nil, fmt.Errorf("clusterCaCertificate not found for cluster %s", args.ClusterPath())
	}
	if endpoint == "" {
		return nil, nil, fmt.Errorf("endpoint not found for cluster %s", args.ClusterPath())
	}

	// Ensure the endpoint starts with "https://"
	if !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}

	// Standard naming convention for gcloud-generated kubeconfigs
	newClusterName := fmt.Sprintf("gke_%s_%s_%s", args.ProjectID, args.Location, args.ClusterName)

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
			&mcp.TextContent{Text: fmt.Sprintf("Kubeconfig for cluster %s (Project: %s, Location: %s) successfully appended/updated in %s. Current context set to %s.", args.ClusterPath(), args.ProjectID, args.Location, pathOptions.GlobalFile, newClusterName)},
		},
	}, nil, nil
}

func (h *handlers) getNodeSosReport(ctx context.Context, _ *mcp.CallToolRequest, args *getNodeSosReportArgs) (res *mcp.CallToolResult, _ any, retErr error) {
	if args == nil || args.Node == "" {
		return nil, nil, fmt.Errorf("node argument cannot be empty")
	}
	// Basic validation for node name to prevent command injection
	if !validNodeRe.MatchString(args.Node) {
		return nil, nil, fmt.Errorf("invalid node name: %s", args.Node)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	destDir, createdTemp, err := resolveSosDestinationDir(cwd, args.Destination)
	if err != nil {
		return nil, nil, err
	}
	if createdTemp {
		defer func() {
			if retErr != nil {
				_ = os.RemoveAll(destDir)
			}
		}()
	}

	method := args.Method
	if method == "" {
		method = "any"
	}
	timeoutSeconds := args.TimeoutSeconds
	if timeoutSeconds <= 0 {
		timeoutSeconds = 180 // Default to 3 minutes
	}

	// Check if node is healthy
	isHealthy := false
	// #nosec G204 -- args.Node is validated against validNodeRe.
	cmd := execCommand(ctx, "kubectl", "get", "node", args.Node, "-o", "jsonpath='{.status.conditions[?(@.type==\"Ready\")].status}'")
	out, err := cmd.Output()
	if err == nil && strings.Contains(string(out), "True") {
		isHealthy = true
	}

	if !isHealthy {
		method = "ssh"
	}

	if method == "pod" || method == "any" {
		// 1. Try Pod-based approach with timeout
		podCtx, podCancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer podCancel()

		res, _, err := h.getNodeSosReportWithPod(podCtx, args.Node, destDir)
		if err == nil {
			return res, nil, nil
		}
		if method == "pod" {
			return nil, nil, fmt.Errorf("failed to get sos report with pod: %w", err)
		}
		// If method is any and pod failed (e.g. timeout), fall through to ssh
	}

	// 2. Fallback or direct SSH approach with timeout
	sshCtx, sshCancel := context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
	defer sshCancel()
	return h.getNodeSosReportWithSSH(sshCtx, args.Node, destDir)
}

func resolveSosDestinationDir(baseDir, rawDest string) (string, bool, error) {
	if rawDest == "" {
		dir, err := os.MkdirTemp("", "gke-sos-report-*")
		if err != nil {
			return "", false, fmt.Errorf("failed to create temporary destination directory: %w", err)
		}
		return dir, true, nil
	}
	if strings.TrimSpace(rawDest) == "" {
		return "", false, fmt.Errorf("destination argument cannot be whitespace")
	}
	if strings.ContainsRune(rawDest, 0) {
		return "", false, fmt.Errorf("destination contains invalid NUL byte")
	}
	// Reject values that could be interpreted as command options or escape
	// the MCP server's working directory.
	if strings.HasPrefix(rawDest, "-") || !filepath.IsLocal(rawDest) {
		return "", false, fmt.Errorf("destination must be a relative path under the current directory: %q", rawDest)
	}

	cleanDest := filepath.Clean(rawDest)
	if strings.HasPrefix(cleanDest, "-") || !filepath.IsLocal(cleanDest) {
		return "", false, fmt.Errorf("destination must be a relative path under the current directory: %q", rawDest)
	}

	if err := ensureWithinBaseDir(baseDir, cleanDest, rawDest); err != nil {
		return "", false, err
	}

	destDir := filepath.Join(baseDir, cleanDest)
	// #nosec G301,G703 -- cleanDest is validated as a local relative path under baseDir without validation-time symlink escapes.
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		return "", false, fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}
	return destDir, false, nil
}

// ensureWithinBaseDir checks that the deepest existing path component does not
// escape baseDir via symlinks at validation time.
func ensureWithinBaseDir(baseDir, relPath, rawDir string) error {
	baseReal, err := filepath.EvalSymlinks(baseDir)
	if err != nil {
		return fmt.Errorf("failed to resolve working directory: %w", err)
	}

	cur := filepath.Join(baseReal, relPath)
	for {
		_, err := os.Lstat(cur)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(cur)
			if err != nil {
				return fmt.Errorf("destination has an unresolvable path component %q: %w", rawDir, err)
			}
			rel, err := filepath.Rel(baseReal, resolved)
			if err != nil || !filepath.IsLocal(rel) {
				return fmt.Errorf("destination must resolve within the current directory: %q", rawDir)
			}
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to inspect destination %q: %w", rawDir, err)
		}
		parent := filepath.Dir(cur)
		if parent == cur || cur == baseReal {
			return fmt.Errorf("failed to inspect destination %q: base directory does not exist", rawDir)
		}
		cur = parent
	}
}

func (h *handlers) getNodeSosReportWithPod(ctx context.Context, node, destDir string) (*mcp.CallToolResult, any, error) {
	// 1. Prepare and run debug pod
	podName := fmt.Sprintf("sos-debug-%d", time.Now().Unix())
	overrides := map[string]interface{}{
		"spec": map[string]interface{}{
			"nodeName":    node,
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
				"kubernetes.io/hostname": node,
			},
		},
	}
	overridesBytes, err := json.Marshal(overrides)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal overrides: %w", err)
	}

	// #nosec G204 -- podName uses current timestamp and node is validated against validNodeRe.
	runCmd := execCommand(ctx, "kubectl", "run", podName, "--image=gke.gcr.io/debian-base", "--restart=Never", "--overrides="+string(overridesBytes))
	if out, err := runCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to create debug pod: %s, %w", string(out), err)
	}

	defer func() {
		// Cleanup pod
		// #nosec G204 -- podName uses server-generated timestamp.
		delCmd := execCommand(context.Background(), "kubectl", "delete", "pod", podName, "--wait=false", "--grace-period=0", "--force")
		_ = delCmd.Run() // Best-effort cleanup
	}()

	// 2. Wait for pod to be ready
	// #nosec G204 -- podName uses server-generated timestamp.
	waitCmd := execCommand(ctx, "kubectl", "wait", "--for=condition=Ready", "pod/"+podName, "--timeout=60s")
	if out, err := waitCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("debug pod did not become ready: %s, %w", string(out), err)
	}

	// 3. Run sos report inside the pod (chrooted to host)
	// We create a temp dir for the report to avoid conflicts in /tmp
	remoteTmpDir := fmt.Sprintf("/tmp/sos-%s", podName)
	// Prepare command: mkdir dir, run sos report, and ensure we capture output
	// Note: chroot /host allows us to use the host's sosreport command and filesystem
	execScript := fmt.Sprintf("apt update && apt install -y sosreport && mkdir -p /host%s && sos report --sysroot=/host --all-logs --batch --tmp-dir=/host%s", remoteTmpDir, remoteTmpDir)

	// #nosec G204 -- podName and remoteTmpDir are server-generated.
	execCmd := execCommand(ctx, "kubectl", "exec", podName, "--", "sh", "-c", execScript)
	outBytes, err := execCmd.CombinedOutput()
	output := string(outBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate sos report: %s, %w", output, err)
	}

	// 4. Parse the output to find the filename
	// The output usually contains: "Your sosreport has been generated and saved in: /path/to/file.tar.xz"
	// The path might be reported as /host/tmp/... or /tmp/... depending on how sos report was invoked.
	// We also handle both .tar.xz and .tar.gz extensions.
	re := regexp.MustCompile(`(/host)?` + regexp.QuoteMeta(remoteTmpDir) + `/[a-zA-Z0-9._-]+\.tar\.(xz|gz)`)
	match := re.FindString(output)
	if match == "" || strings.Contains(match, "..") {
		return nil, nil, fmt.Errorf("could not find sos report filename in output: %s", output)
	}

	// The file path inside the pod is what we need for 'cat'.
	// If the match didn't start with /host, we prepend it because we mounted host root at /host.
	remotePath := match
	if !strings.HasPrefix(remotePath, "/host") {
		remotePath = "/host" + remotePath
	}
	localFilename := fmt.Sprintf("sosreport-%s-%s.tar.xz", node, time.Now().Format("2006-01-02-15-04-05"))
	localPath := filepath.Clean(filepath.Join(destDir, localFilename))
	// 5. Copy the file from the pod to the destination directory
	// #nosec G304,G703 -- destDir is validated within cwd or created via os.MkdirTemp, and localFilename is built from validated node name and timestamp.
	f, err := os.OpenFile(localPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}

	// #nosec G204 -- podName is server-generated and remotePath is matched against a restricted filename regex.
	catCmd := execCommand(ctx, "kubectl", "exec", podName, "--", "cat", "--", remotePath)
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
	// #nosec G204 -- podName and remoteTmpDir are server-generated.
	cleanCmd := execCommand(ctx, "kubectl", "exec", podName, "--", "sh", "-c", cleanupScript)
	_ = cleanCmd.Run() // Best-effort cleanup

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("SOS report successfully generated and downloaded to: %s", localPath)},
		},
	}, nil, nil
}

func (h *handlers) getNodeSosReportWithSSH(ctx context.Context, node, destDir string) (*mcp.CallToolResult, any, error) {
	// 1. Find the zone of the VM
	// gcloud compute instances list --filter="name=NODE_NAME" --format="value(zone)"
	// #nosec G204 -- node is validated against validNodeRe.
	findZoneCmd := execCommand(ctx, "gcloud", "compute", "instances", "list", fmt.Sprintf("--filter=name=%s", node), "--format=value(zone)")
	zoneOut, err := findZoneCmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find zone for node %s using gcloud: %w", node, err)
	}
	zone := strings.TrimSpace(string(zoneOut))
	if zone == "" {
		return nil, nil, fmt.Errorf("could not find zone for node %s", node)
	}
	if !validZoneRe.MatchString(zone) {
		return nil, nil, fmt.Errorf("invalid zone %q for node %s", zone, node)
	}

	// 2. Generate SOS report via SSH
	// gcloud compute ssh --zone "ZONE" "NODE_NAME" --command "sudo sos report --all-logs --batch --tmp-dir=/var"
	// #nosec G204 -- zone and node are validated against validZoneRe and validNodeRe.
	sshCmd := execCommand(ctx, "gcloud", "compute", "ssh", "--zone", zone, node, "--command", "sudo sos report --all-logs --batch --tmp-dir=/var")
	outBytes, err := sshCmd.CombinedOutput()
	output := string(outBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate sos report via ssh: %s, %w", output, err)
	}

	// 3. Parse output for filename
	// Matches /var/sosreport-*.tar.xz
	match := sshSosReportRe.FindString(output)
	if match == "" || strings.Contains(match, "..") {
		return nil, nil, fmt.Errorf("could not find sos report filename in ssh output: %s", output)
	}
	remotePath := match

	// 4. Change ownership of the file
	// gcloud compute ssh ... --command "sudo chown $USER REMOTE_PATH"
	// #nosec G204 -- zone, node, and remotePath are validated against strict regexes.
	chownCmd := execCommand(ctx, "gcloud", "compute", "ssh", "--zone", zone, node, "--command", fmt.Sprintf("sudo chown $USER %s", remotePath))
	if out, err := chownCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to chown remote file: %s, %w", string(out), err)
	}

	// 5. SCP the file
	// gcloud compute scp --zone "ZONE" -- "NODE_NAME:REMOTE_PATH" LOCAL_DESTINATION
	localFilename := fmt.Sprintf("sosreport-%s-%s.tar.xz", node, time.Now().Format("2006-01-02-15-04-05"))
	localPath := filepath.Clean(filepath.Join(destDir, localFilename))
	// #nosec G204 -- zone, node, remotePath, and destDir are validated, and positional args are preceded by "--".
	scpCmd := execCommand(ctx, "gcloud", "compute", "scp", "--zone", zone, "--", fmt.Sprintf("%s:%s", node, remotePath), localPath)
	if out, err := scpCmd.CombinedOutput(); err != nil {
		return nil, nil, fmt.Errorf("failed to scp file: %s, %w", string(out), err)
	}

	// 6. Cleanup remote files on host
	// #nosec G204 -- zone, node, and remotePath are validated against strict regexes.
	rmCmd := execCommand(ctx, "gcloud", "compute", "ssh", "--zone", zone, node, "--command", fmt.Sprintf("sudo rm %s", remotePath))
	_ = rmCmd.Run() // Best-effort cleanup

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("SOS report successfully generated (via SSH) and downloaded to: %s", localPath)},
		},
	}, nil, nil
}

func (h *handlers) updateCluster(ctx context.Context, _ *mcp.CallToolRequest, args *updateClusterArgs) (*mcp.CallToolResult, any, error) {
	var updateObj containerpb.ClusterUpdate
	if err := protojson.Unmarshal([]byte(args.Update), &updateObj); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal update JSON: %w", err)
	}

	req := &containerpb.UpdateClusterRequest{
		Name:   args.ClusterPath(),
		Update: &updateObj,
	}
	resp, err := h.cmClient.UpdateCluster(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) deleteCluster(ctx context.Context, _ *mcp.CallToolRequest, args *deleteClusterArgs) (*mcp.CallToolResult, any, error) {
	if !strings.EqualFold(args.DeletionPolicy, "FORCE") {
		if err := h.verifyClusterUnused(ctx, args.ClusterPath()); err != nil {
			return nil, nil, fmt.Errorf("cluster deletion blocked by safety check: %w. Ensure local kubeconfig is updated (via get_kubeconfig) or set deletionPolicy='FORCE' to override", err)
		}
	}

	req := &containerpb.DeleteClusterRequest{
		Name: args.ClusterPath(),
	}
	resp, err := h.cmClient.DeleteCluster(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: protojson.Format(resp)},
		},
	}, nil, nil
}

func (h *handlers) verifyClusterUnused(ctx context.Context, clusterPath string) error {
	discoveryClient, err := h.k8sProvider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return fmt.Errorf("get discovery client: %w", err)
	}

	dc, err := h.k8sProvider.DynamicClient(ctx, clusterPath)
	if err != nil {
		return fmt.Errorf("get dynamic client: %w", err)
	}

	// 1. External Exposure
	// Presence of >=1 Service of type LoadBalancer, OR any Ingress, Gateway, or MultiClusterIngress.
	if gvr, _, _, err := k8s.ResolveGVR(ctx, discoveryClient, "Service"); err == nil {
		ri := dc.Resource(gvr)
		var limit int64 = 100
		var continueToken string
		for {
			list, err := ri.List(ctx, metav1.ListOptions{Limit: limit, Continue: continueToken})
			if err != nil {
				return fmt.Errorf("list services: %w", err)
			}
			for _, item := range list.Items {
				svcType, found, err := unstructured.NestedString(item.Object, "spec", "type")
				if err == nil && found && svcType == "LoadBalancer" {
					return fmt.Errorf("external exposure: service %s/%s is of type LoadBalancer", item.GetNamespace(), item.GetName())
				}
			}
			continueToken = list.GetContinue()
			if continueToken == "" {
				break
			}
		}
	} else if !isResourceTypeNotFoundError(err) {
		return err
	}

	if err := h.checkAnyResourceExists(ctx, discoveryClient, dc, "Ingress"); err != nil {
		return err
	}

	if err := h.checkAnyResourceExists(ctx, discoveryClient, dc, "Gateway"); err != nil {
		return err
	}

	if err := h.checkAnyResourceExists(ctx, discoveryClient, dc, "MultiClusterIngress"); err != nil {
		return err
	}

	// 2. Persistent Data
	// Presence of >=1 PVC in Bound state in any namespace.
	if gvr, _, _, err := k8s.ResolveGVR(ctx, discoveryClient, "PersistentVolumeClaim"); err == nil {
		ri := dc.Resource(gvr)
		var limit int64 = 100
		var continueToken string
		for {
			list, err := ri.List(ctx, metav1.ListOptions{Limit: limit, Continue: continueToken})
			if err != nil {
				return fmt.Errorf("list pvc: %w", err)
			}
			for _, item := range list.Items {
				phase, found, err := unstructured.NestedString(item.Object, "status", "phase")
				if err == nil && found && phase == "Bound" {
					return fmt.Errorf("persistent data: pvc %s/%s is in bound state", item.GetNamespace(), item.GetName())
				}
			}
			continueToken = list.GetContinue()
			if continueToken == "" {
				break
			}
		}
	} else if !isResourceTypeNotFoundError(err) {
		return err
	}

	// 3. Active Compute
	// Presence of >=1 Pod in Running or Pending state in any user namespace (excluding kube-system or gke-* prefixed).
	if gvr, _, _, err := k8s.ResolveGVR(ctx, discoveryClient, "Pod"); err == nil {
		ri := dc.Resource(gvr)
		var limit int64 = 100
		var continueToken string
		for {
			list, err := ri.List(ctx, metav1.ListOptions{Limit: limit, Continue: continueToken})
			if err != nil {
				return fmt.Errorf("list pods: %w", err)
			}
			for _, item := range list.Items {
				ns := item.GetNamespace()
				if ns == "kube-system" || strings.HasPrefix(ns, "gke-") {
					continue
				}
				phase, found, err := unstructured.NestedString(item.Object, "status", "phase")
				if err == nil && found && (phase == "Running" || phase == "Pending") {
					return fmt.Errorf("active compute: pod %s/%s is in %s state", ns, item.GetName(), phase)
				}
			}
			continueToken = list.GetContinue()
			if continueToken == "" {
				break
			}
		}
	} else if !isResourceTypeNotFoundError(err) {
		return err
	}

	return nil
}

func (h *handlers) checkAnyResourceExists(ctx context.Context, discoveryClient discovery.DiscoveryInterface, dc dynamic.Interface, kind string) error {
	gvr, _, _, err := k8s.ResolveGVR(ctx, discoveryClient, kind)
	if err != nil {
		if isResourceTypeNotFoundError(err) {
			return nil // If the cluster doesn't support the resource type, it's not active
		}
		return fmt.Errorf("resolve %s resource: %w", kind, err)
	}

	ri := dc.Resource(gvr)
	list, err := ri.List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("list %s resources: %w", kind, err)
	}

	if len(list.Items) > 0 {
		item := list.Items[0]
		return fmt.Errorf("external exposure: found active %s resource %s/%s", kind, item.GetNamespace(), item.GetName())
	}

	return nil
}

func isResourceTypeNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	if meta.IsNoMatchError(err) {
		return true
	}
	errStr := err.Error()
	return strings.Contains(errStr, "the server doesn't have a resource type")
}

func prefixStrings(fields []string, prefix string) []string {
	var paths []string
	for _, field := range fields {
		paths = append(paths, prefix+field)
	}
	return paths
}

func initializeDefaultFieldMask(fields []string) string {
	paths := slices.Clone(fields)
	slices.Sort(paths)
	return strings.Join(paths, ",")
}

func getFieldMask(fieldMask string, defaultFieldMask string) string {
	if fieldMask != "" {
		return fieldMask
	}
	return defaultFieldMask
}
