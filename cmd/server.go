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

package main

import (
	"log"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools"
	"github.com/mark3labs/mcp-go/server"
)

const (
	version = "0.0.1"
)

func main() {
	// Create a new MCP server
	s := server.NewMCPServer(
		"GKE MCP Server",
		version,
		server.WithToolCapabilities(true),
	)

	c := config.New(version)
	tools.Install(s, c)

	// Start the stdio server
	log.Printf("Starting GKE MCP Server")
	if err := server.ServeStdio(s); err != nil {
		log.Printf("Server error: %v\n", err)
	}
}
