// Copyright 2026 Google LLC
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

package llm

import (
	"context"
	"os"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
)

func TestNewClient_NilConfig(t *testing.T) {
	_, err := NewClient(context.Background(), nil)

	if err == nil {
		t.Errorf("Expected error for nil config, got nil")
	}
}

func TestNewClient_UnsupportedProvider(t *testing.T) {
	cfg := config.NewTestConfig("test-project", "us-central1", "unsupported-provider", "gemini-2.5-pro")
	_, err := NewClient(context.Background(), cfg)

	if err == nil {
		t.Errorf("Expected error for unsupported provider, got nil")
	}
}

func TestNewClient_CaseInsensitiveProvider(t *testing.T) {
	// Create dummy credentials file
	dummyCreds := `{
	  "type": "service_account",
	  "project_id": "dummy-project",
	  "private_key_id": "12345",
	  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
	  "client_email": "dummy@example.com",
	  "client_id": "12345",
	  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
	  "token_uri": "https://oauth2.googleapis.com/token",
	  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/dummy%40example.com"
	}`
	
	tmpfile, err := os.CreateTemp("", "dummy-creds-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(dummyCreds)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpfile.Name())

	cfg := config.NewTestConfig("test-project", "us-central1", "VeRtEx-AI", "gemini-2.5-pro")
	_, err = NewClient(context.Background(), cfg)

	if err != nil {
		t.Errorf("Expected no error for mixed-case vertex provider with dummy credentials, got %v", err)
	}
}
