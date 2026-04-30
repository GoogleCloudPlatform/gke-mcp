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

package cluster

import (
	"testing"
)

func TestCreateNodePoolArgs_Fields(t *testing.T) {
	args := createNodePoolArgs{}
	args.ProjectID = "test-project"
	args.Location.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.NodePool = `{"name": "my-pool"}`

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.NodePool != `{"name": "my-pool"}` {
		t.Errorf("NodePool = %s, want {\"name\": \"my-pool\"}", args.NodePool)
	}
}

func TestListNodePoolsArgs_Fields(t *testing.T) {
	args := listNodePoolsArgs{}
	args.ProjectID = "test-project"
	args.Location.Location = "us-central1"
	args.ClusterName = "my-cluster"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
}

func TestGetNodePoolArgs_Fields(t *testing.T) {
	args := getNodePoolArgs{}
	args.ProjectID = "test-project"
	args.Location.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.NodePoolName = "my-pool"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.NodePoolName != "my-pool" {
		t.Errorf("NodePoolName = %s, want my-pool", args.NodePoolName)
	}
}

func TestUpdateNodePoolArgs_Fields(t *testing.T) {
	args := updateNodePoolArgs{}
	args.ProjectID = "test-project"
	args.Location.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.NodePoolName = "my-pool"
	args.Update = `{"nodeCount": 5}`

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.NodePoolName != "my-pool" {
		t.Errorf("NodePoolName = %s, want my-pool", args.NodePoolName)
	}
	if args.Update != `{"nodeCount": 5}` {
		t.Errorf("Update = %s, want {\"nodeCount\": 5}", args.Update)
	}
}

func TestDeleteNodePoolArgs_Fields(t *testing.T) {
	args := deleteNodePoolArgs{}
	args.ProjectID = "test-project"
	args.Location.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.NodePoolName = "my-pool"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.NodePoolName != "my-pool" {
		t.Errorf("NodePoolName = %s, want my-pool", args.NodePoolName)
	}
}
