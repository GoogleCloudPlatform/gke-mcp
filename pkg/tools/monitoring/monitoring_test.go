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

package monitoring

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListMonitoredResourceDescriptorsArgs_Fields(t *testing.T) {
	args := listMonitoredResourceDescriptorsArgs{
		ProjectID: "my-project",
	}

	if args.ProjectID != "my-project" {
		t.Errorf("ProjectID = %s, want my-project", args.ProjectID)
	}
}

func TestListMonitoredResourceDescriptorsArgs_Empty(t *testing.T) {
	args := listMonitoredResourceDescriptorsArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
}

func TestListMonitoredResourceDescriptorsArgs_DifferentProjects(t *testing.T) {
	projects := []string{
		"my-project",
		"my-other-project",
		"123456789012",
	}

	for _, project := range projects {
		t.Run(project, func(t *testing.T) {
			args := listMonitoredResourceDescriptorsArgs{
				ProjectID: project,
			}
			if args.ProjectID != project {
				t.Errorf("ProjectID = %s, want %s", args.ProjectID, project)
			}
		})
	}
}

func TestListMonitoredResourceDescriptorsArgs_JSONTags(t *testing.T) {
	args := listMonitoredResourceDescriptorsArgs{
		ProjectID: "test-project",
	}

	if args.ProjectID != "test-project" {
		t.Error("ProjectID field not working correctly")
	}
}

func TestListMonitoredResourceDescriptorsArgs_ZeroValue(t *testing.T) {
	var args listMonitoredResourceDescriptorsArgs
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID for zero value, got %s", args.ProjectID)
	}
}

func TestListMonitoredResourceDescriptorsArgs_WithProjectNumber(t *testing.T) {
	args := listMonitoredResourceDescriptorsArgs{
		ProjectID: "123456789012",
	}

	if args.ProjectID != "123456789012" {
		t.Errorf("ProjectID = %s, want 123456789012", args.ProjectID)
	}
}

func TestQueryPrometheus_ValidationErrors(t *testing.T) {
	ctx := context.Background()

	// Test 1: Empty project ID
	hNoProj := &handlers{
		c: config.NewTestConfig("", "", "test", "test"),
	}
	_, _, err := hNoProj.queryPrometheus(ctx, nil, &queryPrometheusArgs{ProjectID: "", Query: "up"})
	if err == nil || !strings.Contains(err.Error(), "project_id argument cannot be empty") {
		t.Fatalf("Expected project_id error, got: %v", err)
	}

	// Test 2: Empty query
	hWithProj := &handlers{
		c: config.NewTestConfig("test-project", "", "test", "test"),
	}
	_, _, err = hWithProj.queryPrometheus(ctx, nil, &queryPrometheusArgs{ProjectID: "test-project", Query: ""})
	if err == nil || !strings.Contains(err.Error(), "query argument cannot be empty") {
		t.Fatalf("Expected query error, got: %v", err)
	}
}

func TestQueryPrometheus_Success_Base64EncodedData(t *testing.T) {
	ctx := context.Background()

	expectedPromData := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"up"},"value":[1700000000,"1"]}]}}`
	base64Data := base64.StdEncoding.EncodeToString([]byte(expectedPromData))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "projects/test-project/location/global/prometheus/api/v1/query") {
			t.Errorf("Unexpected request URL path: %s", r.URL.Path)
		}
		respMap := map[string]any{
			"contentType": "application/json",
			"data":        base64Data,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respMap)
	}))
	defer ts.Close()

	h := &handlers{
		c:        config.NewTestConfig("test-project", "us-central1", "test", "test"),
		endpoint: ts.URL,
	}

	res, _, err := h.queryPrometheus(ctx, nil, &queryPrometheusArgs{
		ProjectID: "test-project",
		Query:     "up",
	})
	if err != nil {
		t.Fatalf("queryPrometheus returned unexpected error: %v", err)
	}

	if len(res.Content) == 0 {
		t.Fatal("Expected non-empty Content in result")
	}

	textContent, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected *mcp.TextContent, got %T", res.Content[0])
	}

	if textContent.Text != expectedPromData {
		t.Errorf("Got output:\n%s\nWant:\n%s", textContent.Text, expectedPromData)
	}
}

func TestQueryPrometheus_Success_RawMapData(t *testing.T) {
	ctx := context.Background()

	expectedMapData := map[string]any{
		"status": "success",
		"data": map[string]any{
			"resultType": "vector",
			"result":     []any{},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		respMap := map[string]any{
			"contentType": "application/json",
			"data":        expectedMapData,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(respMap)
	}))
	defer ts.Close()

	h := &handlers{
		c:        config.NewTestConfig("test-project", "us-central1", "test", "test"),
		endpoint: ts.URL,
	}

	res, _, err := h.queryPrometheus(ctx, nil, &queryPrometheusArgs{
		ProjectID: "test-project",
		Query:     "up",
	})
	if err != nil {
		t.Fatalf("queryPrometheus returned unexpected error: %v", err)
	}

	if len(res.Content) == 0 {
		t.Fatal("Expected non-empty Content in result")
	}

	textContent, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("Expected *mcp.TextContent, got %T", res.Content[0])
	}

	if !strings.Contains(textContent.Text, `"status":"success"`) {
		t.Errorf("Output does not contain status success: %s", textContent.Text)
	}
}

func TestQueryPrometheus_APIError(t *testing.T) {
	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"code":500,"message":"internal server error"}}`))
	}))
	defer ts.Close()

	h := &handlers{
		c:        config.NewTestConfig("test-project", "us-central1", "test", "test"),
		endpoint: ts.URL,
	}

	_, _, err := h.queryPrometheus(ctx, nil, &queryPrometheusArgs{
		ProjectID: "test-project",
		Query:     "up",
	})
	if err == nil {
		t.Fatal("Expected error for HTTP 500 response, got nil")
	}
	if !strings.Contains(err.Error(), "failed to query prometheus metrics") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestInstall_QueryPrometheusRegistered(t *testing.T) {
	ctx := context.Background()
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "test-server",
		Version: "1.0.0",
	}, nil)
	c := config.NewTestConfig("test-project", "us-central1", "test", "test")

	err := Install(ctx, s, c)
	if err != nil {
		t.Fatalf("Install returned error: %v", err)
	}
}
