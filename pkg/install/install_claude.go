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
	"os/exec"
	"path/filepath"
	"runtime"
)

// ClaudeDesktopExtension installs the GKE MCP Server into Claude Desktop settings
func ClaudeDesktopExtension(opts *InstallOptions) error {
	configPath, err := getClaudeDesktopConfigPath()
	if err != nil {
		return fmt.Errorf("could not determine Claude Desktop config path: %w", err)
	}

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("could not create Claude Desktop config directory: %w", err)
	}

	// Read existing configuration if it exists
	config := make(map[string]interface{})
	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("could not parse existing Claude Desktop config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("could not read Claude Desktop config: %w", err)
	}

	// Add or update the gke-mcp server configuration
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		// Handle the case where mcpServers does not exist or is not a map
		mcpServers = make(map[string]interface{})
		config["mcpServers"] = mcpServers
	}

	mcpServers["gke-mcp"] = map[string]interface{}{
		"command": opts.exePath,
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

// ClaudeCodeExtension installs the GKE MCP Server for Claude Code CLI
func ClaudeCodeExtension(opts *InstallOptions) error {
	installDir := opts.installDir

	// Create the CLAUDE.md file
	claudeMdPath := filepath.Join(installDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMdPath, GeminiMarkdown, 0644); err != nil {
		return fmt.Errorf("could not write CLAUDE.md: %w", err)
	}
	
	// Execute the command to add the MCP server
	command := "claude"
	args := []string{
		"mcp",
		"add",
		"gke-mcp",
		"--command", opts.exePath,
	}

	cmdToRun := exec.Command(command, args...)
	cmdToRun.Stdout = os.Stdout
	cmdToRun.Stderr = os.Stderr

	if err := cmdToRun.Run(); err != nil {
		return fmt.Errorf("failed to run command 'claude mcp add': %w", err)
	}

	return nil
}
