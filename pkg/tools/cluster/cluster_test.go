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
	args := listClustersArgs{
		ProjectID: "test-project",
		Location:  "us-central1",
	}

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
}

func TestGetClustersArgs_Fields(t *testing.T) {
	args := getClustersArgs{
		ProjectID: "test-project",
		Location:  "us-central1",
		Name:      "my-cluster",
	}

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-central1" {
		t.Errorf("Location = %s, want us-central1", args.Location)
	}
	if args.Name != "my-cluster" {
		t.Errorf("Name = %s, want my-cluster", args.Name)
	}
}

func TestGetKubeconfigArgs_Fields(t *testing.T) {
	args := getKubeconfigArgs{
		ProjectID: "test-project",
		Location:  "us-west1",
		Name:      "my-cluster",
	}

	if args.ProjectID != "test-project" {
		t.Errorf("ProjectID = %s, want test-project", args.ProjectID)
	}
	if args.Location != "us-west1" {
		t.Errorf("Location = %s, want us-west1", args.Location)
	}
	if args.Name != "my-cluster" {
		t.Errorf("Name = %s, want my-cluster", args.Name)
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

func TestGetNodeSosReportArgs_Defaults(t *testing.T) {
	args := getNodeSosReportArgs{
		Node: "my-node",
	}

	if args.Destination != "" {
		t.Errorf("Expected empty Destination for default, got %s", args.Destination)
	}
	if args.Method != "" {
		t.Errorf("Expected empty Method for default, got %s", args.Method)
	}
	if args.TimeoutSeconds != 0 {
		t.Errorf("Expected zero TimeoutSeconds for default, got %d", args.TimeoutSeconds)
	}
}

func TestGetNodeSosReportArgs_EmptyNode(t *testing.T) {
	args := getNodeSosReportArgs{}
	if args.Node != "" {
		t.Errorf("Expected empty Node, got %s", args.Node)
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
}

func TestGetClustersArgs_Empty(t *testing.T) {
	args := getClustersArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.Name != "" {
		t.Errorf("Expected empty Name, got %s", args.Name)
	}
}

func TestGetKubeconfigArgs_Empty(t *testing.T) {
	args := getKubeconfigArgs{}
	if args.ProjectID != "" {
		t.Errorf("Expected empty ProjectID, got %s", args.ProjectID)
	}
	if args.Location != "" {
		t.Errorf("Expected empty Location, got %s", args.Location)
	}
	if args.Name != "" {
		t.Errorf("Expected empty Name, got %s", args.Name)
	}
}

func TestListClustersArgs_JSONTags(t *testing.T) {
	args := listClustersArgs{
		ProjectID: "my-project",
		Location:  "us-east1",
	}

	if args.ProjectID != "my-project" {
		t.Error("ProjectID field not working correctly")
	}
}

func TestGetClustersArgs_JSONTags(t *testing.T) {
	args := getClustersArgs{
		ProjectID: "my-project",
		Location:  "us-east1",
		Name:      "my-cluster",
	}

	if args.Name != "my-cluster" {
		t.Error("Name field not working correctly")
	}
}

func TestGetNodeSosReportArgs_JSONTags(t *testing.T) {
	args := getNodeSosReportArgs{
		Node:           "gke-test-pool-abc123",
		Destination:    "/custom/path",
		Method:         "ssh",
		TimeoutSeconds: 600,
	}

	if args.TimeoutSeconds != 600 {
		t.Error("TimeoutSeconds field not working correctly")
	}
}
