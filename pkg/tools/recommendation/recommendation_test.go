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

package recommendation

import (
	"testing"
)

func TestListRecommendationsArgs_Fields(t *testing.T) {
	args := listRecommendationsArgs{
		ProjectID: "my-project",
		Location:  "us-central1",
	}

	if args.ProjectID != "my-project" {
		t.Errorf("ProjectID = %s, want my-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
}

func TestListRecommendationsArgs_Empty(t *testing.T) {
	args := listRecommendationsArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
}

func TestListRecommendationsArgs_DifferentLocations(t *testing.T) {
	locations := []string{
		"us-central1",
		"us-east1",
		"europe-west1",
		"asia-northeast1",
		"-", // all zones
	}

	for _, loc := range locations {
		t.Run(loc, func(t *testing.T) {
			args := listRecommendationsArgs{
				ProjectID: "test-project",
				Location:  loc,
			}
			if args.Location != loc {
				t.Errorf("Location = %s, want %s", args.Location, loc)
			}
		})
	}
}

func TestListRecommendationsArgs_DifferentProjects(t *testing.T) {
	projects := []string{
		"my-project",
		"my-other-project",
		"123456789012",
	}

	for _, project := range projects {
		t.Run(project, func(t *testing.T) {
			args := listRecommendationsArgs{
				ProjectID: project,
				Location:  "us-central1",
			}
			if args.ProjectID != project {
				t.Errorf("ProjectID = %s, want %s", args.ProjectID, project)
			}
		})
	}
}

func TestListRecommendationsArgs_JSONTags(t *testing.T) {
	args := listRecommendationsArgs{
		ProjectID: "test-project",
		Location:  "us-west1",
	}

	if args.ProjectID != "test-project" {
		t.Error("ProjectID field not working correctly")
	}
	if args.Location != "us-west1" {
		t.Error("Location field not working correctly")
	}
}

func TestListRecommendationsArgs_ZeroValue(t *testing.T) {
	var args listRecommendationsArgs
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID for zero value, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location for zero value, got %s", args.Location)
	}
}

func TestListRecommendationsArgs_RegionalCluster(t *testing.T) {
	args := listRecommendationsArgs{
		ProjectID: "my-project",
		Location:  "us-central1",
	}

	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
}

func TestListRecommendationsArgs_ZonalCluster(t *testing.T) {
	args := listRecommendationsArgs{
		ProjectID: "my-project",
		Location:  "us-central1-a",
	}

	if args.Location != "us-central1-a" {
		t.Errorf("Location = %s, want us-central1-a", args.Location)
	}
}

func TestListRecommendationsArgs_AllZones(t *testing.T) {
	args := listRecommendationsArgs{
		ProjectID: "my-project",
		Location:  "-",
	}

	if args.Location != "-" {
		t.Errorf("Location = %s, want -", args.Location)
	}
}
