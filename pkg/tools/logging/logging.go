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

// Package logging provides MCP tools for querying GCP logging.
package logging

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func isMockEnabled() bool {
	exePath, err := os.Executable()
	if err != nil {
		log.Printf("[DEBUG] failed to get executable path: %v", err)
		return false
	}
	exeDir := filepath.Dir(exePath)
	mockFilePath := filepath.Join(exeDir, ".mock_logs_enabled")
	_, err = os.Stat(mockFilePath)
	log.Printf("[DEBUG] checking for mock marker at %s: err=%v", mockFilePath, err)
	return err == nil
}

// Install adds GCP logging related tools to an MCP server.
func Install(_ context.Context, s *mcp.Server, c *config.Config) error {
	realGCP := &gcpLogClient{userAgent: c.UserAgent()}
	var client LogClient
	
	if isMockEnabled() {
		log.Println("[INFO] GKE MCP Logging Tool initialized in MOCK mode (Hybrid).")
		client = &mockLogClient{realClient: realGCP}
	} else {
		client = realGCP
	}

	installQueryLogsTool(s, c, client)
	installGetLogSchemas(s)

	return nil
}
