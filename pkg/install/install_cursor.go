// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CursorMCPConfig represents the MCP configuration for Cursor
type CursorMCPConfig struct {
	MCPServers map[string]CursorMCPServer `json:"mcpServers"`
}

// CursorMCPServer represents an individual MCP server configuration
type CursorMCPServer struct {
	Command string `json:"command"`
	Type    string `json:"type"`
}

// CursorMCPExtension installs the gke-mcp server as a Cursor MCP extension
func CursorMCPExtension(baseDir, exePath string, projectOnlyMode bool) error {
	// Determine the Cursor MCP configuration directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	mcpDir := ""
	if !projectOnlyMode {
		// Create the global Cursor MCP configuration directory
		mcpDir = filepath.Join(homeDir, ".cursor")
	} else {
		// Create project-specific configuration if projectOnlyMode set to true
		mcpDir = filepath.Join(baseDir, ".cursor")
	}
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("could not create Cursor directory at %s: %w", mcpDir, err)
	}
	mcpPath := filepath.Join(mcpDir, "mcp.json")

	// Read existing configuration if it exists
	config := CursorMCPConfig{
		MCPServers: make(map[string]CursorMCPServer),
	}

	if _, err := os.Stat(mcpPath); err == nil {
		// File exists, read and parse it
		data, err := os.ReadFile(mcpPath)
		if err != nil {
			return fmt.Errorf("could not read existing MCP configuration: %w", err)
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("could not parse existing MCP configuration: %w", err)
		}
	}

	// Add or update the gke-mcp server configuration
	config.MCPServers["gke-mcp"] = CursorMCPServer{
		Command: exePath,
		Type:    "stdio",
	}

	// Write the updated configuration back to the file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal MCP configuration: %w", err)
	}

	if err := os.WriteFile(mcpPath, data, 0644); err != nil {
		return fmt.Errorf("could not write MCP configuration: %w", err)
	}

	// Create the rules directory and gke-mcp.mdc file
	rulesDir := filepath.Join(mcpDir, "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		return fmt.Errorf("could not create rules directory: %w", err)
	}

	// Read the GEMINI.md content
	geminiContent, err := os.ReadFile(filepath.Join(baseDir, "pkg", "install", "GEMINI.md"))
	if err != nil {
		return fmt.Errorf("could not read GEMINI.md file: %w", err)
	}

	// Create the gke-mcp.mdc rule file with custom heading and GEMINI.md content
	ruleContent := `---
name: GKE MCP Instructions
description: Provides guidance for using the gke-mcp tool with Cursor.
alwaysApply: true
---

# GKE MCP Tool Instructions

This rule provides context for using the gke-mcp tool within Cursor.

` + string(geminiContent)

	rulePath := filepath.Join(rulesDir, "gke-mcp.mdc")
	if err := os.WriteFile(rulePath, []byte(ruleContent), 0644); err != nil {
		return fmt.Errorf("could not write gke-mcp rule file: %w", err)
	}

	return nil
}
