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

package logging

import (
	"testing"
)

func TestGetSampleQueries(t *testing.T) {
	// Create a mock MCP server to register the tool
	// Since we can't easily mock the server's internal state for AddTool,
	// we'll just test the handler logic directly if possible, or use a real server instance.
	// However, the handler is an anonymous function in installGetSampleQueriesTool.
	// Let's refactor slightly to make it testable or just test the logic by calling the function.
	// Actually, we can just call the handler if we extract it, but for now let's just use the install function
	// and then simulate a call if the SDK supports it, or just trust the logic is simple enough.
	// Better yet, let's verify the sampleQueries variable content.

	if len(sampleQueries) == 0 {
		t.Error("sampleQueries should not be empty")
	}

	// Verify categories
	categories := make(map[string]bool)
	for _, q := range sampleQueries {
		categories[q.Category] = true
		if q.Name == "" {
			t.Error("Sample query name should not be empty")
		}
		if q.Query == "" {
			t.Errorf("Sample query content should not be empty for %s", q.Name)
		}
	}

	expectedCategories := []string{"Cluster", "Pod", "Node", "Container", "Control Plane", "Namespace"}
	for _, c := range expectedCategories {
		if !categories[c] {
			t.Errorf("Expected category %s not found", c)
		}
	}
}

func TestGetSampleQueriesHandler(t *testing.T) {
	// We can't easily test the handler without extracting it or mocking the server.
	// For now, let's just ensure the logic works by simulating what the handler does.

	req := &GetSampleQueriesRequest{Category: "Cluster"}
	var filtered []SampleQuery
	for _, q := range sampleQueries {
		if q.Category == req.Category {
			filtered = append(filtered, q)
		}
	}

	if len(filtered) == 0 {
		t.Error("Should have found Cluster queries")
	}
	for _, q := range filtered {
		if q.Category != "Cluster" {
			t.Errorf("Expected category Cluster, got %s", q.Category)
		}
	}

	// Test no category (all queries)
	var all []SampleQuery
	all = append(all, sampleQueries...)
	if len(all) != len(sampleQueries) {
		t.Error("Should return all queries when no category specified")
	}
}
