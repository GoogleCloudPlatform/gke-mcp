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
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed GEMINI.md
var GeminiMarkdown []byte

func GeminiCLIExtension(baseDir, version, exePath string) error {
	extensionDir := filepath.Join(baseDir, ".gemini", "extensions", "gke-mcp")
	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		return fmt.Errorf("could not create extension directory: %w", err)
	}

	// Create the manifest file as described in https://github.com/google-gemini/gemini-cli/blob/main/docs/extension.md.
	manifest := map[string]interface{}{
		"name":            "gke-mcp",
		"version":         version,
		"description":     "Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.",
		"contextFileName": "GEMINI.md",
		"mcpServers": map[string]interface{}{
			"gke": map[string]interface{}{
				"command": exePath,
			},
		},
	}

	manifestPath := filepath.Join(extensionDir, "gemini-extension.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal manifest.json: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("could not write manifest.json: %w", err)
	}

	geminiMdPath := filepath.Join(extensionDir, "GEMINI.md")
	if err := os.WriteFile(geminiMdPath, GeminiMarkdown, 0644); err != nil {
		return fmt.Errorf("could not write GEMINI.md: %w", err)
	}

	return nil
}
