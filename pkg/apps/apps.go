// Copyright 2026 Google LLC
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

// Package apps registers MCP apps for GKE workflows.
package apps

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/apps/dropdown"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type installer func(ctx context.Context, s *mcp.Server, c *config.Config) error

// InstallApps registers MCP tools that require a client host with 'apps' extension support.
func InstallApps(ctx context.Context, s *mcp.Server, c *config.Config) error {
	installers := []installer{
		dropdown.Install,
	}

	for _, installer := range installers {
		if err := installer(ctx, s, c); err != nil {
			return err
		}
	}

	return nil
}
