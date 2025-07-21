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
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGeminiCLIExtension(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gemini-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testVersion := "0.1.0-test"
	testExePath := "/usr/local/bin/gke-mcp"
	if err := GeminiCLIExtension(tmpDir, testVersion, testExePath, false); err != nil {
		t.Fatalf("GeminiCLIExtension() failed: %v", err)
	}

	extensionDir := filepath.Join(tmpDir, ".gemini", "extensions", "gke-mcp")
	manifestPath := filepath.Join(extensionDir, "gemini-extension.json")
	geminiMdPath := filepath.Join(extensionDir, "GEMINI.md")

	if _, err := os.Stat(geminiMdPath); os.IsNotExist(err) {
		t.Errorf("Expected GEMINI.md file to be created at %s, but it was not", geminiMdPath)
	}

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest file: %v", err)
	}

	expectedJSON := `{
  "name": "gke-mcp",
  "version": "0.1.0-test",
  "description": "Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.",
  "contextFileName": "GEMINI.md",
  "mcpServers": {
    "gke": {
      "command": "/usr/local/bin/gke-mcp"
    }
  }
}`

	var actual, expected map[string]interface{}
	if err := json.Unmarshal(manifestData, &actual); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}
	if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
		t.Fatalf("Failed to unmarshal expected JSON: %v", err)
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("Manifest content mismatch. Diff:\n%v", diff)
	}
}

func TestGeminiCLIExtensionDeveloperMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gemini-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	geminiMdPath := filepath.Join(tmpDir, "pkg", "install", "GEMINI.md")
	if err := os.MkdirAll(filepath.Dir(geminiMdPath), 0700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(geminiMdPath, GeminiMarkdown, 0600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	testVersion := "0.1.0-test"
	testExePath := filepath.Join(tmpDir, "gke-mcp")
	if err := GeminiCLIExtension(tmpDir, testVersion, testExePath, true); err != nil {
		t.Fatalf("GeminiCLIExtension() failed: %v", err)
	}

	extensionDir := filepath.Join(tmpDir, ".gemini", "extensions", "gke-mcp")
	manifestPath := filepath.Join(extensionDir, "gemini-extension.json")

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest file: %v", err)
	}

	expected := map[string]any{
		"name":            "gke-mcp",
		"version":         "0.1.0-test",
		"description":     "Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.",
		"contextFileName": geminiMdPath,
		"mcpServers": map[string]any{
			"gke": map[string]any{
				"command": testExePath,
			},
		},
	}

	var actual map[string]any
	if err := json.Unmarshal(manifestData, &actual); err != nil {
		t.Fatalf("Failed to unmarshal actual JSON: %v", err)
	}

	if diff := cmp.Diff(actual, expected); diff != "" {
		t.Errorf("Manifest content mismatch. Diff:\n%v", diff)
	}
}
