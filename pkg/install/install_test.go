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
)

func TestGeminiCLIExtension(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gemini-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testVersion := "0.1.0-test"
	testExePath := "/usr/local/bin/gke-mcp"
	if err := GeminiCLIExtension(tmpDir, testVersion, testExePath); err != nil {
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

	actualCanonical, _ := json.MarshalIndent(actual, "", "  ")
	expectedCanonical, _ := json.MarshalIndent(expected, "", "  ")

	if string(actualCanonical) != string(expectedCanonical) {
		t.Errorf("Manifest content mismatch.\nGot:\n%s\n\nExpected:\n%s", string(manifestData), expectedJSON)
	}
}
