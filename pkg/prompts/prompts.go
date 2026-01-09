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

package prompts

import (
	"context"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/prompts/cost"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/prompts/deploy"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/prompts/upgraderiskreport"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/prompts/upgradesbestpracticesriskreport"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type installer func(ctx context.Context, s *mcp.Server, c *config.Config) error

func Install(ctx context.Context, s *mcp.Server, c *config.Config) error {
	installers := []installer{
		cost.Install,
		deploy.Install,
		upgraderiskreport.Install,
		upgradesbestpracticesriskreport.Install,
	}

	for _, installer := range installers {
		if err := installer(ctx, s, c); err != nil {
			return err
		}
	}

	return nil
}
