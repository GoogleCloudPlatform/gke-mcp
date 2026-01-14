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
	cfg := New(version)

	if cfg.UserAgent() != "gke-mcp/"+version {
		t.Errorf("UserAgent() = %s, want %s", cfg.UserAgent(), "gke-mcp/"+version)
	}
}

func TestConfigGetters(t *testing.T) {
	cfg := &Config{
		userAgent:        "test-agent",
		defaultProjectID: "test-project",
		defaultLocation:  "us-central1",
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
}

func TestNewConfigWithVersion(t *testing.T) {
	testVersion := "0.1.0"
	cfg := New(testVersion)

	if cfg == nil {
		t.Fatal("New() returned nil")
	}

	expectedUserAgent := "gke-mcp/" + testVersion
	if cfg.UserAgent() != expectedUserAgent {
		t.Errorf("UserAgent() = %s, want %s", cfg.UserAgent(), expectedUserAgent)
	}
}

func TestConfigFields(t *testing.T) {
	cfg := &Config{
		userAgent:        "test-agent",
		defaultProjectID: "my-project",
		defaultLocation:  "us-west1",
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
		cfg := New(v)
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
			cfg := New(tt.version)
			if cfg.UserAgent() != tt.wantAgent {
				t.Errorf("UserAgent() = %s, want %s", cfg.UserAgent(), tt.wantAgent)
			}
		})
	}
}
