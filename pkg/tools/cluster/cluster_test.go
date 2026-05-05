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

func TestListClustersArgs_Fields(t *testing.T) {
	args := listClustersArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ReadMask = "name,status"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
}

func TestGetClustersArgs_Fields(t *testing.T) {
	args := getClustersArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.ReadMask = "name,status"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}

	if args.ReadMask != "name,status" {
		t.Errorf("ReadMask = %s, want name,status", args.ReadMask)
	}
}

func TestCreateClustersArgs_Fields(t *testing.T) {
	args := createClustersArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.Cluster = `{"name": "my-cluster"}`

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}

	if args.Cluster != `{"name": "my-cluster"}` {
		t.Errorf("Cluster = %s, want {\"name\": \"my-cluster\"}", args.Cluster)
	}
}

func TestGetKubeconfigArgs_Fields(t *testing.T) {
	var args getKubeconfigArgs
	args.ProjectID = "test-project"
	args.Location = "us-west1"
	args.ClusterName = "my-cluster"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-west1" {
		t.Errorf("Location = %s, want us-west1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
}

func TestGetNodeSosReportArgs_Fields(t *testing.T) {
	args := getNodeSosReportArgs{
		Node:           "my-node",
		Destination:    "/tmp/sos",
		Method:         "pod",
		TimeoutSeconds: 300,
	}

	if args.Node != "my-node" {
		t.Errorf("Node = %s, want my-node", args.Node)
	}
	if args.Destination != "/tmp/sos" {
		t.Errorf("Destination = %s, want /tmp/sos", args.Destination)
	}
	if args.Method != "pod" {
		t.Errorf("Method = %s, want pod", args.Method)
	}
	if args.TimeoutSeconds != 300 {
		t.Errorf("TimeoutSeconds = %d, want 300", args.TimeoutSeconds)
	}
}

func TestListClustersArgs_Empty(t *testing.T) {
	args := listClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.ReadMask != "" {
		t.Errorf("Expected empty ReadMask, got %s", args.ReadMask)
	}
	if args.ReadMask != "" {
		t.Errorf("Expected empty ReadMask, got %s", args.ReadMask)
	}
}

func TestGetClustersArgs_Empty(t *testing.T) {
	args := getClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.ClusterName != "" {
		t.Errorf("Expected empty ClusterName, got %s", args.ClusterName)
	}
	if args.ReadMask != "" {
		t.Errorf("Expected empty ReadMask, got %s", args.ReadMask)
	}
}

func TestCreateClustersArgs_Empty(t *testing.T) {
	args := createClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.Cluster != "" {
		t.Errorf("Expected empty Cluster, got %s", args.Cluster)
	}
}

func TestUpdateClusterArgs_Fields(t *testing.T) {
	args := updateClusterArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ClusterName = "my-cluster"
	args.Update = `{"description": "new description"}`

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
	if args.Update != `{"description": "new description"}` {
		t.Errorf("Update = %s, want {\"description\": \"new description\"}", args.Update)
	}
}

func TestDeleteClusterArgs_Fields(t *testing.T) {
	args := deleteClusterArgs{}
	args.ProjectID = "test-project"
	args.Location = "us-central1"
	args.ClusterName = "my-cluster"

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.ClusterName != "my-cluster" {
		t.Errorf("ClusterName = %s, want my-cluster", args.ClusterName)
	}
}
