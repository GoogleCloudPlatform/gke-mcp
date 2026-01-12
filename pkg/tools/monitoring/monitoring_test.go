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
	"testing"
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
