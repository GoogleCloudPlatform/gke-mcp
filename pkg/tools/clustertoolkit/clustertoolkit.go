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

package clustertoolkit

import (
	"context"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Install(_ context.Context, s *server.MCPServer, _ *config.Config) error {
	clusterToolkitDownloadTool := mcp.NewTool("cluster_toolkit_download",
		mcp.WithDescription("Cluster Toolkit, is open-source software offered by Google Cloud which simplifies the process for you to create Google Kubernetes Engine clusters and deploy high performance computing (HPC), artificial intelligence (AI), and machine learning (ML). It is designed to be highly customizable and extensible, and intends to address the deployment needs of a broad range of use cases. This tool will download the public git repository so that Cluster Toolkit can be used."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("download_directory", mcp.Required(), mcp.Description("Download directory for the git repo. By default use the absolute path to the current working directory.")),
	)
	s.AddTool(clusterToolkitDownloadTool, clusterToolkitDownload)

	return nil
}

func clusterToolkitDownload(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	download_dir, err := request.RequireString("download_directory")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	// Make sure we download into a sub-directory
	if !strings.HasSuffix(download_dir, "cluster-toolkit") {
		download_dir = filepath.Join(download_dir, "cluster-toolkit")
	}
	out, err := exec.Command("git", "clone", "https://github.com/GoogleCloudPlatform/cluster-toolkit.git", download_dir).Output()
	if err != nil {
		log.Printf("Failed to download Cluster Toolkit: %v %s", err, out)
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(string(out)), nil
}
