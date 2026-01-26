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
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// testSetup creates a temporary directory and optionally mocks the HOME environment
func testSetup(t *testing.T, mockHome bool) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tmpDir) // Best-effort cleanup
	}

	if mockHome {
		originalHome := os.Getenv("HOME")
		_ = os.Setenv("HOME", tmpDir) // Test environment setup
		cleanup = func() {
			_ = os.RemoveAll(tmpDir)            // Best-effort cleanup
			_ = os.Setenv("HOME", originalHome) // Restore environment
		}
	}

	return tmpDir, cleanup
}

// mockAppData mocks the APPDATA environment variables to a temporary directory
// for the duration of a test. It returns a cleanup function to restore the original values.
func mockAppData(_ *testing.T, tmpDir string) func() {
	originalAppData := os.Getenv("APPDATA")

	if runtime.GOOS == "windows" {
		_ = os.Setenv("APPDATA", tmpDir) // Test environment setup
	}

	return func() {
		_ = os.Setenv("APPDATA", originalAppData) // Restore environment
	}
}

// mockInput simulates user input for interactive prompts
func mockInput(input string) func() {
	// Create a pipe to simulate user input
	r, w, _ := os.Pipe()
	oldStdin := os.Stdin
	os.Stdin = r

	// Write the input to the pipe
	go func() {
		defer func() { _ = w.Close() }()
		_, _ = w.WriteString(input) // Test input
	}()

	// Return cleanup function
	return func() {
		os.Stdin = oldStdin
		_ = r.Close() // Cleanup pipe
	}
}

// MockClaudeCommand creates a mock 'claude' command in a temporary directory
// and sets up the PATH environment variable so that it is found before the real command.
// It returns the log file path which includes its arguments.
func MockClaudeCommand(t *testing.T) (logFile string, cleanup func()) {
	tmpDir, err := os.MkdirTemp("", "claude-mock")
	if err != nil {
		t.Fatalf("Failed to create temp dir for mock claude: %v", err)
	}

	logFile = filepath.Join(tmpDir, "claude-log.txt")
	claudePath := filepath.Join(tmpDir, "claude")
	if runtime.GOOS == "windows" {
		claudePath += ".bat"
	}

	var mockScript string
	if runtime.GOOS == "windows" {
		// On Windows, %* represents all arguments
		mockScript = fmt.Sprintf("@echo off\necho %%* >> %s\n", logFile)
	} else {
		// On Unix/Linux/macOS, "$@" represents all arguments
		mockScript = fmt.Sprintf("#!/bin/bash\necho \"$@\" >> %s\n", logFile)
	}

	// #nosec G306
	if err := os.WriteFile(claudePath, []byte(mockScript), 0700); err != nil {
		t.Fatalf("Failed to create logging claude command: %v", err)
	}

	originalPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+originalPath)

	cleanup = func() {
		_ = os.Setenv("PATH", originalPath) // Restore environment
		_ = os.RemoveAll(tmpDir)            // Best-effort cleanup
	}

	return logFile, cleanup
}

// Verifies that the expected arguments are in the log file.
func verifyArgs(t *testing.T, logFile string, testExePath string) {

	// Verify the claude command was called with correct arguments
	// #nosec G304
	logContent, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read command log: %v", err)
	}

	expectedArgs := fmt.Sprintf("mcp add gke-mcp %s", testExePath)
	if !strings.Contains(string(logContent), expectedArgs) {
		t.Errorf("Expected claude command to be called with args '%s', but log contains: %s", expectedArgs, string(logContent))
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

// verifyClaudeCodeInstallation checks that Claude Code installation created the expected files
func verifyClaudeCodeInstallation(t *testing.T, installDir string, _ string) {
	claudeMDPath := filepath.Join(installDir, "CLAUDE.md")
	usageGuidePath := filepath.Join(installDir, "GKE_MCP_USAGE_GUIDE.md")

	// Verify CLAUDE.md exists and has correct content
	if _, err := os.Stat(claudeMDPath); os.IsNotExist(err) {
		t.Errorf("Expected CLAUDE.md file to be created at %s, but it was not", claudeMDPath)
	} else {
		// #nosec G304
		claudeContent, err := os.ReadFile(claudeMDPath)
		if err != nil {
			t.Fatalf("Failed to read CLAUDE.md: %v", err)
		}

		expectedReference := fmt.Sprintf("# GKE-MCP Server Instructions\n - @%s", usageGuidePath)
		if !strings.Contains(string(claudeContent), expectedReference) {
			t.Errorf("Expected CLAUDE.md to contain reference to usage guide, but it was not found.\nContent: %s\nExpected: %s", string(claudeContent), expectedReference)
		}
	}

	// Verify GKE_MCP_USAGE_GUIDE.md exists and has correct content
	if _, err := os.Stat(usageGuidePath); os.IsNotExist(err) {
		t.Errorf("Expected GKE_MCP_USAGE_GUIDE.md file to be created at %s, but it was not", usageGuidePath)
	} else {
		// #nosec G304
		usageContent, err := os.ReadFile(usageGuidePath)
		if err != nil {
			t.Fatalf("Failed to read GKE_MCP_USAGE_GUIDE.md: %v", err)
		}

		// Verify content matches GeminiMarkdown
		if !bytes.Equal(usageContent, GeminiMarkdown) {
			t.Errorf("Expected GKE_MCP_USAGE_GUIDE.md content to match GeminiMarkdown")
		}
	}
}

// verifyMCPConfig validates the MCP configuration file content
func verifyMCPConfig(t *testing.T, mcpPath, expectedExePath string) {
	// #nosec G304
	// #nosec G304
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
	if err := os.MkdirAll(cursorDir, 0700); err != nil {
		t.Fatalf("Failed to create cursor directory: %v", err)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	mcpPath := filepath.Join(cursorDir, "mcp.json")
	if err := os.WriteFile(mcpPath, configData, 0600); err != nil {
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
	// #nosec G304
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
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testVersion := "0.1.0-test"
	testExePath := "/usr/local/bin/gke-mcp"
	opts := &Options{
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

	// #nosec G304
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
	defer func() { _ = os.RemoveAll(tmpDir) }()

	geminiMdPath := filepath.Join(tmpDir, "pkg", "install", "GEMINI.md")
	if err := os.MkdirAll(filepath.Dir(geminiMdPath), 0700); err != nil {
		t.Fatalf("os.MkdirAll() failed: %v", err)
	}
	if err := os.WriteFile(geminiMdPath, GeminiMarkdown, 0600); err != nil {
		t.Fatalf("os.WriteFile() failed: %v", err)
	}

	testVersion := "0.1.0-test"
	testExePath := filepath.Join(tmpDir, "gke-mcp")
	opts := &Options{
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
	// #nosec G304
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
	opts := &Options{
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
	opts := &Options{
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

	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := CursorMCPExtension(opts); err != nil {
		t.Fatalf("CursorMCPExtension() failed: %v", err)
	}

	// Verify that existing configuration is preserved
	// #nosec G304
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
	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := CursorMCPExtension(opts); err != nil {
		t.Fatalf("CursorMCPExtension() failed: %v", err)
	}

	// Verify that the malformed config was fixed
	// #nosec G304
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

// verifyGkeMcpInClaudeConfig checks for the presence and correctness of the gke-mcp server entry.
func verifyGkeMcpInClaudeConfig(t *testing.T, config map[string]interface{}, expectedExePath string) {
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

// verifyClaudeDesktopConfig validates the Claude Desktop configuration file content
func verifyClaudeDesktopConfig(t *testing.T, configPath, expectedExePath string) {
	// #nosec G304
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Claude Desktop config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal Claude Desktop config: %v", err)
	}

	verifyGkeMcpInClaudeConfig(t, config, expectedExePath)
}

// createExistingClaudeConfig creates a pre-existing Claude Desktop configuration file for testing
func createExistingClaudeConfig(t *testing.T, configPath string, config map[string]interface{}) {
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		t.Fatalf("Failed to create mock config directory: %v", err)
	}

	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0600); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
}

func TestClaudeDesktopExtension(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"

	// Mock the config path by temporarily setting environment variables
	cleanupEnv := mockAppData(t, tmpDir)
	defer cleanupEnv()

	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := ClaudeDesktopExtension(opts); err != nil {
		t.Fatalf("ClaudeDesktopExtension() failed: %v", err)
	}

	expectedConfigPath, err := getClaudeDesktopConfigPath()
	if err != nil {
		t.Fatalf("could not determine Claude Desktop config path: %v", err)
	}
	verifyClaudeDesktopConfig(t, expectedConfigPath, testExePath)
}

func TestClaudeDesktopExtensionWithExistingConfig(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	// Mock environment variables
	cleanupEnv := mockAppData(t, tmpDir)
	defer cleanupEnv()

	// Create existing Claude Desktop configuration
	configPath, err := getClaudeDesktopConfigPath()
	if err != nil {
		t.Fatalf("could not determine Claude Desktop config path: %v", err)
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

	createExistingClaudeConfig(t, configPath, existingConfig)

	// Install gke-mcp
	testExePath := "/usr/local/bin/gke-mcp"
	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := ClaudeDesktopExtension(opts); err != nil {
		t.Fatalf("ClaudeDesktopExtension() failed: %v", err)
	}

	// Verify that existing configuration is preserved
	// #nosec G304
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
	verifyGkeMcpInClaudeConfig(t, config, testExePath)
}

func TestClaudeDesktopExtensionWithMalformedConfig(t *testing.T) {
	tmpDir, cleanup := testSetup(t, true)
	defer cleanup()

	// Mock environment variables
	cleanupEnv := mockAppData(t, tmpDir)
	defer cleanupEnv()

	// Create malformed Claude Desktop configuration (mcpServers as string instead of map)
	configPath, err := getClaudeDesktopConfigPath()
	if err != nil {
		t.Fatalf("could not determine Claude Desktop config path: %v", err)
	}

	malformedConfig := map[string]interface{}{
		"mcpServers":   "this should be a map, not a string",
		"otherSetting": "value",
	}

	createExistingClaudeConfig(t, configPath, malformedConfig)

	// Install gke-mcp - this should handle the malformed config gracefully
	testExePath := "/usr/local/bin/gke-mcp"

	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}
	if err := ClaudeDesktopExtension(opts); err != nil {
		t.Fatalf("ClaudeDesktopExtension() failed: %v", err)
	}

	// Verify that the malformed config was fixed
	// #nosec G304
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read Claude Desktop config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		t.Fatalf("Failed to unmarshal Claude Desktop config: %v", err)
	}

	verifyGkeMcpInClaudeConfig(t, config, testExePath)
}

// Claude Code Extension Tests

func TestClaudeCodeExtension(t *testing.T) {
	tmpDir, cleanup := testSetup(t, false)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"

	logFile, cleanupCommand := MockClaudeCommand(t)
	defer cleanupCommand()

	// Mock user input to answer "yes" to the confirmation prompt
	cleanupInput := mockInput("yes\n")
	defer cleanupInput()

	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}

	if err := ClaudeCodeExtension(opts); err != nil {
		t.Fatalf("ClaudeCodeExtension() failed: %v", err)
	}

	// Verify installation
	verifyClaudeCodeInstallation(t, tmpDir, testExePath)

	verifyArgs(t, logFile, testExePath)
}

func TestClaudeCodeExtensionWithExistingClaude(t *testing.T) {
	tmpDir, cleanup := testSetup(t, false)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"

	// Create existing CLAUDE.md file
	claudeMDPath := filepath.Join(tmpDir, "CLAUDE.md")
	existingContent := "# Existing Content\nSome existing instructions."
	if err := os.WriteFile(claudeMDPath, []byte(existingContent), 0600); err != nil {
		t.Fatalf("Failed to create existing CLAUDE.md: %v", err)
	}

	// Mock Claude Command
	logFile, cleanupCommand := MockClaudeCommand(t)
	defer cleanupCommand()

	// Mock user input to answer "yes" to the confirmation prompt
	cleanupInput := mockInput("yes\n")
	defer cleanupInput()

	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}

	if err := ClaudeCodeExtension(opts); err != nil {
		t.Fatalf("ClaudeCodeExtension() failed: %v", err)
	}

	// Verify that existing content is preserved and new content is appended
	// #nosec G304
	claudeContent, err := os.ReadFile(claudeMDPath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	// Should contain both existing content and new reference
	if !strings.Contains(string(claudeContent), existingContent) {
		t.Errorf("Expected CLAUDE.md to preserve existing content")
	}

	// Verify other files were created
	verifyClaudeCodeInstallation(t, tmpDir, testExePath)

	//verify claude command execution
	verifyArgs(t, logFile, testExePath)
}

func TestClaudeCodeExtensionUserDeclines(t *testing.T) {
	tmpDir, cleanup := testSetup(t, false)
	defer cleanup()

	testExePath := "/usr/local/bin/gke-mcp"

	// Mock user input to answer "no" to the confirmation prompt
	cleanupInput := mockInput("no\n")
	defer cleanupInput()

	opts := &Options{
		installDir: tmpDir,
		exePath:    testExePath,
	}

	// This should not return an error, but should not create files
	if err := ClaudeCodeExtension(opts); err != nil {
		t.Fatalf("ClaudeCodeExtension() failed: %v", err)
	}

	// Verify that files were NOT created
	claudeMDPath := filepath.Join(tmpDir, "CLAUDE.md")
	usageGuidePath := filepath.Join(tmpDir, "GKE_MCP_USAGE_GUIDE.md")

	if _, err := os.Stat(claudeMDPath); err == nil {
		t.Errorf("Expected CLAUDE.md to NOT be created when user declines, but it was")
	}

	if _, err := os.Stat(usageGuidePath); err == nil {
		t.Errorf("Expected GKE_MCP_USAGE_GUIDE.md to NOT be created when user declines, but it was")
	}
}
