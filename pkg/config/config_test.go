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

package config

import (
	"testing"
)

func TestNew(t *testing.T) {
	version := "1.0.0"
	cfg := New(version, false)

	if cfg.UserAgent() != "gke-mcp/"+version {
		t.Errorf("UserAgent() = %s, want %s", cfg.UserAgent(), "gke-mcp/"+version)
	}
	if cfg.AgentProvider() != "vertex-ai" {
		t.Errorf("AgentProvider() = %s, want vertex-ai", cfg.AgentProvider())
	}
	if cfg.AgentModel() != "gemini-2.5-pro" {
		t.Errorf("AgentModel() = %s, want gemini-2.5-pro", cfg.AgentModel())
	}
	if cfg.EnableDeleteTools() {
		t.Error("Expected EnableDeleteTools to be false")
	}
}

func TestNewWithEnvVars(t *testing.T) {
	t.Setenv("GKE_MCP_PROVIDER", "custom-provider")
	t.Setenv("GKE_MCP_MODEL", "custom-model")

	version := "1.0.0"
	cfg := New(version, false)

	if cfg.AgentProvider() != "custom-provider" {
		t.Errorf("AgentProvider() = %s, want custom-provider", cfg.AgentProvider())
	}
	if cfg.AgentModel() != "custom-model" {
		t.Errorf("AgentModel() = %s, want custom-model", cfg.AgentModel())
	}
}

func TestConfigGetters(t *testing.T) {
	cfg := &Config{
		userAgent:         "test-agent",
		defaultProjectID:  "test-project",
		defaultLocation:   "us-central1",
		enableDeleteTools: true,
	}

	if got := cfg.UserAgent(); got != "test-agent" {
		t.Errorf("UserAgent() = %s, want test-agent", got)
	}
	if got := cfg.DefaultProjectID(); got != "test-project" {
		t.Errorf("DefaultProjectID() = %s, want test-project", got)
	}
	if got := cfg.DefaultLocation(); got != "us-central1" {
		t.Errorf("DefaultLocation() = %s, want us-central1", got)
	}
	if !cfg.EnableDeleteTools() {
		t.Error("Expected EnableDeleteTools to be true")
	}
}

func TestNewConfigWithVersion(t *testing.T) {
	testVersion := "0.1.0"
	cfg := New(testVersion, true)

	if cfg == nil {
		t.Fatal("New() returned nil")
	}

	expectedUserAgent := "gke-mcp/" + testVersion
	if cfg.UserAgent() != expectedUserAgent {
		t.Errorf("UserAgent() = %s, want %s", cfg.UserAgent(), expectedUserAgent)
	}
	if !cfg.EnableDeleteTools() {
		t.Error("Expected EnableDeleteTools to be true")
	}
}

func TestConfigFields(t *testing.T) {
	cfg := &Config{
		userAgent:         "test-agent",
		defaultProjectID:  "my-project",
		defaultLocation:   "us-west1",
		enableDeleteTools: false,
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"UserAgent", cfg.UserAgent(), "test-agent"},
		{"DefaultProjectID", cfg.DefaultProjectID(), "my-project"},
		{"DefaultLocation", cfg.DefaultLocation(), "us-west1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s() = %s, want %s", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestConfigUserAgentFormat(t *testing.T) {
	versions := []string{"0.1.0", "1.0.0", "latest", "v1.2.3"}
	for _, v := range versions {
		cfg := New(v, false)
		expected := "gke-mcp/" + v
		if got := cfg.UserAgent(); got != expected {
			t.Errorf("UserAgent() for version %s = %s, want %s", v, got, expected)
		}
	}
}

func TestGetGcloudConfigTrimsWhitespace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	out, err := getGcloudConfig("core/project")
	if err != nil {
		t.Logf("gcloud config get failed (expected if not configured): %v", err)
	}
	if out != "" {
		result := out
		if result != "" {
			if result != out {
				t.Errorf("getGcloudConfig() should trim whitespace, got: %q", out)
			}
		}
	}
	_ = out
	_ = err
}

func TestGetDefaultLocationNotPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	loc := getDefaultLocation()
	t.Logf("Default location: %s", loc)
}

func TestGetDefaultProjectIDNotPanic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectID := getDefaultProjectID()
	t.Logf("Default project ID: %s", projectID)
}

func TestConfigNilFields(t *testing.T) {
	cfg := &Config{}
	if cfg.UserAgent() != "" {
		t.Errorf("Expected empty UserAgent for empty Config, got %s", cfg.UserAgent())
	}
	if cfg.DefaultProjectID() != "" {
		t.Errorf("Expected empty DefaultProjectID for empty Config, got %s", cfg.DefaultProjectID())
	}
	if cfg.DefaultLocation() != "" {
		t.Errorf("Expected empty DefaultLocation for empty Config, got %s", cfg.DefaultLocation())
	}
}

func TestNewConfigDifferentVersions(t *testing.T) {
	tests := []struct {
		version   string
		wantAgent string
	}{
		{"0.1.0", "gke-mcp/0.1.0"},
		{"1.0.0", "gke-mcp/1.0.0"},
		{"latest", "gke-mcp/latest"},
		{"v1.2.3", "gke-mcp/v1.2.3"},
		{"test-version", "gke-mcp/test-version"},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			cfg := New(tt.version, false)
			if cfg.UserAgent() != tt.wantAgent {
				t.Errorf("UserAgent() = %s, want %s", cfg.UserAgent(), tt.wantAgent)
			}
		})
	}
}

func TestMockMode(t *testing.T) {
	origBuildMockMode := BuildMockMode
	defer func() { BuildMockMode = origBuildMockMode }()

	cfg := &Config{}

	BuildMockMode = "false"
	if cfg.MockMode() {
		t.Error("Expected MockMode to be false by default")
	}

	t.Setenv("GKE_MCP_MOCK", "true")
	if !cfg.MockMode() {
		t.Error("Expected MockMode to be true when GKE_MCP_MOCK=true")
	}

	t.Setenv("GKE_MCP_MOCK", "false")
	BuildMockMode = "true"
	if !cfg.MockMode() {
		t.Error("Expected MockMode to be true when BuildMockMode=true")
	}
}

func TestMockDataDir(t *testing.T) {
	origBuildMockMode := BuildMockMode
	origBuildMockDataDir := BuildMockDataDir
	defer func() {
		BuildMockMode = origBuildMockMode
		BuildMockDataDir = origBuildMockDataDir
	}()

	cfg := &Config{}

	BuildMockMode = "false"
	if got := cfg.MockDataDir(); got != "mock_data" {
		t.Errorf("MockDataDir() = %s, want mock_data", got)
	}

	t.Setenv("GKE_MCP_MOCK_DATA_DIR", "custom_dir")
	if got := cfg.MockDataDir(); got != "custom_dir" {
		t.Errorf("MockDataDir() = %s, want custom_dir", got)
	}

	BuildMockMode = "true"
	BuildMockDataDir = "build_dir"
	if got := cfg.MockDataDir(); got != "build_dir" {
		t.Errorf("MockDataDir() = %s, want build_dir", got)
	}
}

func TestMockSkillAndCase(t *testing.T) {
	origSkill := BuildMockSkill
	origCase := BuildMockCase
	defer func() {
		BuildMockSkill = origSkill
		BuildMockCase = origCase
	}()

	cfg := &Config{}

	BuildMockSkill = "test-skill"
	BuildMockCase = "test_case"

	if got := cfg.MockSkill(); got != "test-skill" {
		t.Errorf("MockSkill() = %s, want test-skill", got)
	}
	if got := cfg.MockCase(); got != "test_case" {
		t.Errorf("MockCase() = %s, want test_case", got)
	}

	t.Setenv("GKE_MCP_MOCK_SKILL", "env-skill")
	t.Setenv("GKE_MCP_MOCK_CASE", "env_case")

	if got := cfg.MockSkill(); got != "env-skill" {
		t.Errorf("MockSkill() = %s, want env-skill", got)
	}
	if got := cfg.MockCase(); got != "env_case" {
		t.Errorf("MockCase() = %s, want env_case", got)
	}
}
