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

// Package clustertoolkit provides tools for downloading Cluster Toolkit.
package clustertoolkit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

const (
	clusterToolkitRepoURL = "https://github.com/GoogleCloudPlatform/cluster-toolkit.git"
	clusterToolkitDirName = "cluster-toolkit"
)

var execCommand = exec.CommandContext

type clusterToolkitDownloadArgs struct {
	DownloadDirectory string `json:"download_directory" jsonschema:"Relative path within the MCP server's current working directory to download the git repo into. Use '.' for the server's current working directory."`
}

// Install registers Cluster Toolkit tools with the MCP server.
func Install(_ context.Context, s *mcp.Server, _ *config.Config) error {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "cluster_toolkit_download",
		Description: "Cluster Toolkit, is open-source software offered by Google Cloud which simplifies the process for you to create Google Kubernetes Engine clusters and deploy high performance computing (HPC), artificial intelligence (AI), and machine learning (ML). It is designed to be highly customizable and extensible, and intends to address the deployment needs of a broad range of use cases. This tool will download the public git repository so that Cluster Toolkit can be used.",
	}, clusterToolkitDownload)

	return nil
}

func clusterToolkitDownload(ctx context.Context, _ *mcp.CallToolRequest, args *clusterToolkitDownloadArgs) (*mcp.CallToolResult, any, error) {
	if args == nil || strings.TrimSpace(args.DownloadDirectory) == "" {
		return nil, nil, fmt.Errorf("download_directory argument cannot be empty")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	downloadDir, err := resolveDownloadDir(cwd, args.DownloadDirectory)
	if err != nil {
		return nil, nil, err
	}

	destPath := filepath.Join(cwd, downloadDir)
	_, statErr := os.Lstat(destPath)
	existedBefore := !errors.Is(statErr, os.ErrNotExist)

	// #nosec G204 -- downloadDir is validated as a local relative path under cwd and preceded by "--".
	cmd := execCommand(ctx, "git", "clone", clusterToolkitRepoURL, "--", downloadDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if !existedBefore {
			_ = os.RemoveAll(destPath)
		}
		outStr := strings.TrimSpace(string(out))
		log.Printf("Failed to download Cluster Toolkit: %v %s", err, outStr)
		if outStr != "" {
			return nil, nil, fmt.Errorf("failed to download Cluster Toolkit: %w: %s", err, outStr)
		}
		return nil, nil, fmt.Errorf("failed to download Cluster Toolkit: %w", err)
	}

	msg := fmt.Sprintf("Downloaded Cluster Toolkit to %s", destPath)
	if outStr := strings.TrimSpace(string(out)); outStr != "" {
		msg = fmt.Sprintf("%s\n%s", msg, outStr)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, nil, nil
}

func resolveDownloadDir(baseDir, rawDir string) (string, error) {
	if strings.TrimSpace(rawDir) == "" {
		return "", fmt.Errorf("download_directory argument cannot be empty")
	}
	if strings.ContainsRune(rawDir, 0) {
		return "", fmt.Errorf("download_directory contains invalid NUL byte")
	}
	// Reject values that could be interpreted as git options or escape the
	// current working directory (absolute paths, parent directory traversal,
	// or platform-reserved names).
	if strings.HasPrefix(rawDir, "-") || !filepath.IsLocal(rawDir) {
		return "", fmt.Errorf("download_directory must be a relative path under the current directory: %q", rawDir)
	}

	cleanDir := filepath.Clean(rawDir)
	if strings.HasPrefix(cleanDir, "-") || !filepath.IsLocal(cleanDir) {
		return "", fmt.Errorf("download_directory must be a relative path under the current directory: %q", rawDir)
	}

	// Make sure we download into a sub-directory.
	downloadDir := cleanDir
	if filepath.Base(cleanDir) != clusterToolkitDirName {
		downloadDir = filepath.Join(cleanDir, clusterToolkitDirName)
	}

	if err := ensureWithinBaseDir(baseDir, downloadDir, rawDir); err != nil {
		return "", err
	}

	return downloadDir, nil
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
				return fmt.Errorf("download_directory has an unresolvable path component %q: %w", rawDir, err)
			}
			rel, err := filepath.Rel(baseReal, resolved)
			if err != nil || !filepath.IsLocal(rel) {
				return fmt.Errorf("download_directory must resolve within the current directory: %q", rawDir)
			}
			return nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to inspect download_directory %q: %w", rawDir, err)
		}
		parent := filepath.Dir(cur)
		if parent == cur || cur == baseReal {
			return fmt.Errorf("failed to inspect download_directory %q: base directory does not exist", rawDir)
		}
		cur = parent
	}
}
