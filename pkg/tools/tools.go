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

// Package tools registers MCP tools for GKE workflows.
package tools

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/cluster"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/clustertoolkit"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/deploy"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/giq"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/gkereleasenotes"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/k8schangelog"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/logging"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/monitoring"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/recommendation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type installer func(ctx context.Context, s *mcp.Server, c *config.Config) error

// Install registers all tools with the MCP server.
func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	installers := []installer{
		cluster.Install,
		clustertoolkit.Install,
		deploy.Install,
		giq.Install,
		logging.Install,
		monitoring.Install,
		recommendation.Install,
		k8schangelog.Install,
		gkereleasenotes.Install,
	}

	for _, installer := range installers {
		if err := installer(ctx, s, c); err != nil {
			return err
		}
	}

	return nil
}
