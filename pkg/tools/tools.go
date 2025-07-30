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

package tools

import (
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/cluster"
	cluster_toolkit "github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/cluster-toolkit"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/giq"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/logging"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/memrag"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/recommendation"
	"github.com/mark3labs/mcp-go/server"
)

func Install(s *server.MCPServer, c *config.Config) {
	cluster.Install(s, c)
	cluster_toolkit.Install(s, c)
	giq.Install(s, c)
	logging.Install(s, c)
	memrag.Install(s, c)
	recommendation.Install(s, c)
}
