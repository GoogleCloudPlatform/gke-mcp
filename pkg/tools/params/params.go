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

// Package params provide common tool parameter types.
package params

import "fmt"

type Project struct {
	ProjectID string `json:"project_id" jsonschema:"Required. GCP project ID."`
}

func (p *Project) ProjectIDPath() string {
	return fmt.Sprintf("projects/%s", p.ProjectID)
}

type LocationRequired struct {
	Project
	Location string `json:"location" jsonschema:"Required to be a valid GCP region or zone. MUST NOT be empty."`
}

func (l *LocationRequired) LocationPath() string {
	return fmt.Sprintf("%s/locations/%s", l.ProjectIDPath(), l.Location)
}

type LocationOptional struct {
	Project
	Location string `json:"location,omitempty" jsonschema:"Optional. GCP region or zone."`
}

func (l *LocationOptional) LocationPath() string {
	if l.Location == "" {
		return fmt.Sprintf("%s/locations/-", l.ProjectIDPath())
	}
	return fmt.Sprintf("%s/locations/%s", l.ProjectIDPath(), l.Location)
}

type Cluster struct {
	LocationRequired
	ClusterName string `json:"cluster_name" jsonschema:"Required. GKE cluster name."`
}

func (c *Cluster) ClusterPath() string {
	return fmt.Sprintf("%s/clusters/%s", c.LocationPath(), c.ClusterName)
}

// NodePool represents GKE node pool parameters.
type NodePool struct {
	Cluster
	NodePoolName string `json:"node_pool_name" jsonschema:"Required. GKE node pool name."`
}

// NodePoolPath returns the full resource path for the node pool.
func (n *NodePool) NodePoolPath() string {
	return fmt.Sprintf("%s/nodePools/%s", n.ClusterPath(), n.NodePoolName)
}
