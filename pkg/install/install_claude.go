// pkg/install/claude.go
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
	"runtime"
)

// ClaudeDesktopConfig represents the structure of claude_desktop_config.json
type ClaudeDesktopConfig struct {
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// MCPServerConfig represents a single MCP server configuration
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// ClaudeDesktopExtension installs the GKE MCP Server into Claude Desktop settings
func ClaudeDesktopExtension(version, exePath string, developerMode bool) error {
	configPath, err := getClaudeDesktopConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine Claude Desktop config path: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("could not create Claude Desktop config directory: %w", err)
	}

	// Read existing config or create new one
	config := &ClaudeDesktopConfig{
		MCPServers: make(map[string]MCPServerConfig),
	}

	if _, err := os.Stat(configPath); err == nil {
		// Config file exists, read it
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("could not read existing Claude Desktop config: %w", err)
		}

		if err := json.Unmarshal(data, config); err != nil {
			return fmt.Errorf("could not parse existing Claude Desktop config: %w", err)
		}
	}

	// Add or update the gke-mcp server configuration
	config.MCPServers["gke-mcp"] = MCPServerConfig{
		Command: exePath,
		Env:     make(map[string]string), // Initialize empty env map
	}

	// Write the updated config back
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal Claude Desktop config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("could not write Claude Desktop config: %w", err)
	}

	return nil
}

// getClaudeDesktopConfigPath returns the platform-specific path to Claude Desktop's config file
func getClaudeDesktopConfigPath() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "darwin": // macOS
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, "Library", "Application Support", "Claude")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		configDir = filepath.Join(appData, "Claude")
	case "linux":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".config", "Claude")
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	return filepath.Join(configDir, "claude_desktop_config.json"), nil
}