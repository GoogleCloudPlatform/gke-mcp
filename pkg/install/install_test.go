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
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// testSetup creates a temporary directory and optionally mocks the HOME environment
func testSetup(t *testing.T, mockHome bool) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "cursor-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	if mockHome {
		originalHome := os.Getenv("HOME")
		os.Setenv("HOME", tmpDir)
		cleanup = func() {
			os.RemoveAll(tmpDir)
			os.Setenv("HOME", originalHome)
		}
	}

	return tmpDir, cleanup
}

// mockHomeAndAppData mocks the HOME and APPDATA environment variables to a temporary directory
// for the duration of a test. It returns a cleanup function to restore the original values.
func mockHomeAndAppData(t *testing.T, tmpDir string) func() {
	originalHome := os.Getenv("HOME")
	originalAppData := os.Getenv("APPDATA")

	os.Setenv("HOME", tmpDir)
	if runtime.GOOS == "windows" {
		os.Setenv("APPDATA", tmpDir)
	}

	return func() {
		os.Setenv("HOME", originalHome)
		os.Setenv("APPDATA", originalAppData)
	}
}

// verifyCursorInstallation checks that the Cursor installation created the expected files and structure
func verifyCursorInstallation(t *testing.T, baseDir string, projectOnly bool) {
	// Determine expected paths
	var mcpPath, rulesPath string
	if projectOnly {
		mcpPath = filepath.Join(baseDir, ".cursor", "mcp.json")
		rulesPath = filepath.Join(baseDir, ".cursor", "rules", "gke-mcp.mdc")
	} else {
		// For global installation, we need to check the mocked home directory
		mcpPath = filepath.Join(baseDir, ".cursor", "mcp.json")
		rulesPath = filepath.Join(baseDir, ".cursor", "rules", "gke-mcp.mdc")
	}

	// Verify MCP configuration file exists
	if _, err := os.Stat(mcpPath); os.IsNotExist(err) {
		t.Errorf("Expected MCP config file to be created at %s, but it was not", mcpPath)
	}

	// Verify rules file exists and has correct content
	verifyRuleFile(t, rulesPath)
}

// verifyMCPConfig validates the MCP configuration file content
func verifyMCPConfig(t *testing.T, mcpPath, expectedExePath string) {
	mcpData, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("Failed to read MCP config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(mcpData, &config); err != nil {
		t.Fatalf("Failed to unmarshal MCP config: %v", err)
	}

	// Verify mcpServers exists and is a map
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mcpServers to be a map, got %T", config["mcpServers"])
	}

	// Verify gke-mcp server configuration
	gkeMcp, ok := mcpServers["gke-mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected gke-mcp to be a map, got %T", mcpServers["gke-mcp"])
	}

	if gkeMcp["command"] != expectedExePath {
		t.Errorf("Expected command to be %s, got %v", expectedExePath, gkeMcp["command"])
	}

	if gkeMcp["type"] != "stdio" {
		t.Errorf("Expected type to be 'stdio', got %v", gkeMcp["type"])
	}
}

// createExistingConfig creates a pre-existing MCP configuration file for testing and returns the path to the file
func createExistingConfig(t *testing.T, cursorDir string, config map[string]interface{}) string {
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatalf("Failed to create cursor directory: %v", err)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	mcpPath := filepath.Join(cursorDir, "mcp.json")
	if err := os.WriteFile(mcpPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return mcpPath
}

// verifyRuleFile checks that the rule file was created with the correct content
func verifyRuleFile(t *testing.T, rulesPath string) {
	if _, err := os.Stat(rulesPath); os.IsNotExist(err) {
		t.Errorf("Expected rules file to be created at %s, but it was not", rulesPath)
		return
	}

	// Read and verify the rule file content
	ruleData, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("Failed to read rule file: %v", err)
	}

	ruleContent := string(ruleData)

	// Check that the header is present using the constant from the package
	if !strings.Contains(ruleContent, cursorRuleHeader) {
		t.Errorf("Expected rule file to contain header, but it was not found")
	}

	// Check that GeminiMarkdown content is present
	if !strings.Contains(ruleContent, string(GeminiMarkdown)) {
		t.Errorf("Expected rule file to contain GeminiMarkdown content, but it was not found")
	}
}

func TestGeminiCLIExtension(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gemini-cli-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testVersion := "0.1.0-test"
	testExePath := "/usr/local/bin/gke-mcp"
	opts := &InstallOptions{
		version:       testVersion,
		installDir:    tmpDir,
		exePath:       testExePath,
		developerMode: false,
	}

	if err := GeminiCLIExtension(opts); err != nil {
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
	tmpDir, err := os.MkdirTemp(".", ".gemini-cli-test")
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
	opts := &InstallOptions{
		version:       testVersion,
		installDir:    tmpDir,
		exePath:       testExePath,
		developerMode: true,
	}

	if err := GeminiCLIExtension(opts); err != nil {
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

func TestCursorMCPExtensionGlobal(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"
	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}

	if err := CursorMCPExtension(opts); err != nil {
		t.Fatalf("CursorMCPExtension() failed: %v", err)
	}

	verifyCursorInstallation(t, tmpDir, false)
	verifyMCPConfig(t, filepath.Join(tmpDir, ".cursor", "mcp.json"), testExePath)
}

func TestCursorMCPExtensionProjectOnly(t *testing.T) {
	tmpDir, cleanup := testSetup(t, false)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"
	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := CursorMCPExtension(opts); err != nil {
		t.Fatalf("CursorMCPExtension() failed: %v", err)
	}

	verifyCursorInstallation(t, tmpDir, true)
	verifyMCPConfig(t, filepath.Join(tmpDir, ".cursor", "mcp.json"), testExePath)
}

func TestCursorMCPExtensionWithExistingConfig(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	// Create existing MCP configuration
	cursorDir := filepath.Join(tmpDir, ".cursor")
	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"existing-server": map[string]interface{}{
				"command": "/usr/bin/existing",
				"type":    "stdio",
			},
		},
		"otherSetting": "value",
	}

	mcpPath := createExistingConfig(t, cursorDir, existingConfig)

	// Install gke-mcp
	testExePath := "/usr/local/bin/gke-mcp"

	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := CursorMCPExtension(opts); err != nil {
		t.Fatalf("CursorMCPExtension() failed: %v", err)
	}

	// Verify that existing configuration is preserved
	mcpData, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("Failed to read MCP config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(mcpData, &config); err != nil {
		t.Fatalf("Failed to unmarshal MCP config: %v", err)
	}

	// Check that existing server is preserved
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mcpServers to be a map, got %T", config["mcpServers"])
	}

	existingServer, ok := mcpServers["existing-server"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected existing-server to be preserved, got %T", mcpServers["existing-server"])
	}

	if existingServer["command"] != "/usr/bin/existing" {
		t.Errorf("Expected existing server command to be preserved, got %v", existingServer["command"])
	}

	// Check that other settings are preserved
	if config["otherSetting"] != "value" {
		t.Errorf("Expected otherSetting to be preserved, got %v", config["otherSetting"])
	}

	// Check that gke-mcp was added
	gkeMcp, ok := mcpServers["gke-mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected gke-mcp to be added, got %T", mcpServers["gke-mcp"])
	}

	if gkeMcp["command"] != testExePath {
		t.Errorf("Expected gke-mcp command to be %s, got %v", testExePath, gkeMcp["command"])
	}
}

func TestCursorMCPExtensionWithMalformedConfig(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	// Create malformed MCP configuration (mcpServers as string instead of map)
	cursorDir := filepath.Join(tmpDir, ".cursor")
	malformedConfig := map[string]interface{}{
		"mcpServers":   "this should be a map, not a string",
		"otherSetting": "value",
	}

	mcpPath := createExistingConfig(t, cursorDir, malformedConfig)

	// Install gke-mcp - this should handle the malformed config gracefully
	testExePath := "/usr/local/bin/gke-mcp"
	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := CursorMCPExtension(opts); err != nil {
		t.Fatalf("CursorMCPExtension() failed: %v", err)
	}

	// Verify that the malformed config was fixed
	mcpData, err := os.ReadFile(mcpPath)
	if err != nil {
		t.Fatalf("Failed to read MCP config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(mcpData, &config); err != nil {
		t.Fatalf("Failed to unmarshal MCP config: %v", err)
	}

	// Check that mcpServers is now a proper map
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mcpServers to be fixed and become a map, got %T", config["mcpServers"])
	}

	// Check that gke-mcp was added successfully
	gkeMcp, ok := mcpServers["gke-mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected gke-mcp to be added, got %T", mcpServers["gke-mcp"])
	}

	if gkeMcp["command"] != testExePath {
		t.Errorf("Expected gke-mcp command to be %s, got %v", testExePath, gkeMcp["command"])
	}
}

// verifyClaudeDesktopConfig validates the Claude Desktop configuration file content
func verifyClaudeDesktopConfig(t *testing.T, configPath, expectedExePath string) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Claude Desktop config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal Claude Desktop config: %v", err)
	}

	// Verify mcpServers exists and is a map
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mcpServers to be a map, got %T", config["mcpServers"])
	}

	// Verify gke-mcp server configuration
	gkeMcp, ok := mcpServers["gke-mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected gke-mcp to be a map, got %T", mcpServers["gke-mcp"])
	}

	if gkeMcp["command"] != expectedExePath {
		t.Errorf("Expected command to be %s, got %v", expectedExePath, gkeMcp["command"])
	}
}

// createExistingClaudeConfig creates a pre-existing Claude Desktop configuration file for testing
func createExistingClaudeConfig(t *testing.T, configDir string, config map[string]interface{}) string {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create claude config directory: %v", err)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(configDir, "claude_desktop_config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	return configPath
}

// mockClaudeDesktopConfigPath mocks the getClaudeDesktopConfigPath function for testing
func mockClaudeDesktopConfigPath(tmpDir string) string {
	var configDir string
	switch runtime.GOOS {
	case "darwin": // macOS
		configDir = filepath.Join(tmpDir, "Library", "Application Support", "Claude")
	case "windows":
		configDir = filepath.Join(tmpDir, "Claude")
	case "linux":
		configDir = filepath.Join(tmpDir, ".config", "Claude")
	default:
        // For unsupported OS, fall back to linux-style path for testing
		// In production, this would return an error
		configDir = filepath.Join(tmpDir, ".config", "Claude")
	}
	return filepath.Join(configDir, "claude_desktop_config.json")
}

func TestClaudeDesktopExtension(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"

	// Mock the config path by temporarily setting environment variables
	cleanupEnv := mockHomeAndAppData(t, tmpDir)
    defer cleanupEnv()

	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := ClaudeDesktopExtension(opts); err != nil {
		t.Fatalf("ClaudeDesktopExtension() failed: %v", err)
	}

	expectedConfigPath := mockClaudeDesktopConfigPath(tmpDir)
	verifyClaudeDesktopConfig(t, expectedConfigPath, testExePath)
}

func TestClaudeDesktopExtensionWithExistingConfig(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	// Mock environment variables
	cleanupEnv := mockHomeAndAppData(t, tmpDir)
    defer cleanupEnv()

	// Create existing Claude Desktop configuration
	var configDir string
	switch runtime.GOOS {
	case "darwin":
		configDir = filepath.Join(tmpDir, "Library", "Application Support", "Claude")
	case "windows":
		configDir = filepath.Join(tmpDir, "Claude")
	default:
		configDir = filepath.Join(tmpDir, ".config", "Claude")
	}

	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"existing-server": map[string]interface{}{
				"command": "/usr/bin/existing",
				"env":     map[string]interface{}{},
			},
		},
		"otherSetting": "value",
	}

	configPath := createExistingClaudeConfig(t, configDir, existingConfig)

	// Install gke-mcp
	testExePath := "/usr/local/bin/gke-mcp"
	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := ClaudeDesktopExtension(opts); err != nil {
		t.Fatalf("ClaudeDesktopExtension() failed: %v", err)
	}

	// Verify that existing configuration is preserved
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Claude Desktop config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal Claude Desktop config: %v", err)
	}

	// Check that existing server is preserved
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mcpServers to be a map, got %T", config["mcpServers"])
	}

	existingServer, ok := mcpServers["existing-server"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected existing-server to be preserved, got %T", mcpServers["existing-server"])
	}

	if existingServer["command"] != "/usr/bin/existing" {
		t.Errorf("Expected existing server command to be preserved, got %v", existingServer["command"])
	}

	// Check that other settings are preserved
	if config["otherSetting"] != "value" {
		t.Errorf("Expected otherSetting to be preserved, got %v", config["otherSetting"])
	}

	// Check that gke-mcp was added
	gkeMcp, ok := mcpServers["gke-mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected gke-mcp to be added, got %T", mcpServers["gke-mcp"])
	}

	if gkeMcp["command"] != testExePath {
		t.Errorf("Expected gke-mcp command to be %s, got %v", testExePath, gkeMcp["command"])
	}
}

func TestClaudeDesktopExtensionWithMalformedConfig(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	// Mock environment variables
	cleanupEnv := mockHomeAndAppData(t, tmpDir)
    defer cleanupEnv()

	// Create malformed Claude Desktop configuration (mcpServers as string instead of map)
	var configDir string
	switch runtime.GOOS {
	case "darwin":
		configDir = filepath.Join(tmpDir, "Library", "Application Support", "Claude")
	case "windows":
		configDir = filepath.Join(tmpDir, "Claude")
	default:
		configDir = filepath.Join(tmpDir, ".config", "Claude")
	}

	malformedConfig := map[string]interface{}{
		"mcpServers":   "this should be a map, not a string",
		"otherSetting": "value",
	}

	configPath := createExistingClaudeConfig(t, configDir, malformedConfig)

	// Install gke-mcp - this should handle the malformed config gracefully
	testExePath := "/usr/local/bin/gke-mcp"

	opts := &InstallOptions{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := ClaudeDesktopExtension(opts); err != nil {
		t.Fatalf("ClaudeDesktopExtension() failed: %v", err)
	}

	// Verify that the malformed config was fixed
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Claude Desktop config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal Claude Desktop config: %v", err)
	}

	// Check that mcpServers is now a proper map
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected mcpServers to be fixed and become a map, got %T", config["mcpServers"])
	}

	// Check that gke-mcp was added successfully
	gkeMcp, ok := mcpServers["gke-mcp"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected gke-mcp to be added, got %T", mcpServers["gke-mcp"])
	}

	if gkeMcp["command"] != testExePath {
		t.Errorf("Expected gke-mcp command to be %s, got %v", testExePath, gkeMcp["command"])
	}
}

