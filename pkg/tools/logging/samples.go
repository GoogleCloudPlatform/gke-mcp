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
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type SampleQuery struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Category    string `json:"category"`
}

var sampleQueries = []SampleQuery{
	// Cluster-level queries
	{
		Name:        "Cluster Activity",
		Description: "General cluster activity logs",
		Query:       `resource.type="gke_cluster" AND log_id("cloudaudit.googleapis.com/activity")`,
		Category:    "Cluster",
	},
	{
		Name:        "Cluster Creation",
		Description: "Logs for cluster creation events",
		Query:       `resource.type="gke_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.methodName="google.container.v1.ClusterManager.CreateCluster"`,
		Category:    "Cluster",
	},
	{
		Name:        "Deployments",
		Description: "Logs related to deployments",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.methodName:"deployments"`,
		Category:    "Cluster",
	},
	{
		Name:        "Anonymous Access",
		Description: "Logs for actions performed by system:anonymous",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.authenticationInfo.principalEmail="system:anonymous"`,
		Category:    "Cluster",
	},
	{
		Name:        "Location Filter",
		Description: "Logs filtered by location (example: us-central1-b)",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="us-central1-b"`,
		Category:    "Cluster",
	},
	{
		Name:        "User Pod Access",
		Description: "Logs for pod access by a specific user",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.methodName:"io.k8s.core.v1.pods" AND protoPayload.authenticationInfo.principalEmail="USER_EMAIL"`,
		Category:    "Cluster",
	},
	{
		Name:        "Cluster Events",
		Description: "General cluster events",
		Query:       `resource.type="k8s_cluster" AND log_id("events")`,
		Category:    "Cluster",
	},
	{
		Name:        "Endpoint Changes",
		Description: "Logs for changes to Endpoints",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.request.kind="Endpoints"`,
		Category:    "Cluster",
	},
	{
		Name:        "K8s Service Activity",
		Description: "Logs for k8s.io service activity",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.serviceName="k8s.io"`,
		Category:    "Cluster",
	},
	{
		Name:        "Container Service Activity",
		Description: "Logs for container.googleapis.com service activity",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.serviceName="container.googleapis.com"`,
		Category:    "Cluster",
	},
	{
		Name:        "Pod Create/Delete",
		Description: "Logs for pod creation and deletion",
		Query:       `resource.type="k8s_cluster" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.methodName=~"io\.k8s\.core\.v1\.pods\.(create|delete)"`,
		Category:    "Cluster",
	},
	{
		Name:        "Pod Resource Activity",
		Description: "Logs for specific pod resource activity",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.resourceName="core/v1/namespaces/POD_NAMESPACE/pods/POD_NAME"`,
		Category:    "Cluster",
	},
	{
		Name:        "Pod Eviction",
		Description: "Logs for pod eviction creation",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.methodName="io.k8s.core.v1.pods.eviction.create"`,
		Category:    "Cluster",
	},
	{
		Name:        "Node Activity",
		Description: "Logs for node activity",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.methodName:"io.k8s.core.v1.nodes"`,
		Category:    "Cluster",
	},
	{
		Name:        "Addon Manager Activity",
		Description: "Logs for actions performed by system:addon-manager",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.authenticationInfo.principalEmail="system:addon-manager"`,
		Category:    "Cluster",
	},
	{
		Name:        "Non-Conflict Errors",
		Description: "Logs for errors that are not conflicts",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("cloudaudit.googleapis.com/activity") AND protoPayload.status.message!="Conflict" AND protoPayload.status.code!=0`,
		Category:    "Cluster",
	},
	{
		Name:        "LoadBalancer Controller Events",
		Description: "Events from the loadbalancer controller",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("events") AND jsonPayload.source.component="loadbalancer-controller"`,
		Category:    "Cluster",
	},
	{
		Name:        "Service Controller Events",
		Description: "Events from the service controller",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("events") AND jsonPayload.source.component="service-controller"`,
		Category:    "Cluster",
	},
	{
		Name:        "Cluster Autoscaler Events",
		Description: "Events from the cluster autoscaler",
		Query:       `resource.type="k8s_cluster" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("events") AND jsonPayload.source.component="cluster-autoscaler"`,
		Category:    "Cluster",
	},

	// Pod-level queries
	{
		Name:        "Pod Events",
		Description: "Events for a specific pod",
		Query:       `resource.type="k8s_pod" AND resource.labels.pod_name="POD_NAME" AND log_id("events")`,
		Category:    "Pod",
	},
	{
		Name:        "Pod Evicted",
		Description: "Logs for evicted pods",
		Query:       `resource.type="k8s_pod" AND log_id("events") AND jsonPayload.reason="Evicted"`,
		Category:    "Pod",
	},
	{
		Name:        "Scheduler Events",
		Description: "Events from the default scheduler",
		Query:       `resource.type="k8s_pod" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("events") AND jsonPayload.source.component="default-scheduler"`,
		Category:    "Pod",
	},
	{
		Name:        "Preempted Pods",
		Description: "Logs for preempted pods",
		Query:       `resource.type="k8s_pod" AND resource.labels.location="CLUSTER_LOCATION" AND resource.labels.cluster_name="CLUSTER_NAME" AND log_id("events") AND jsonPayload.source.component="default-scheduler" AND jsonPayload.reason="Preempted"`,
		Category:    "Pod",
	},

	// Node-level queries
	{
		Name:        "Node Events",
		Description: "Events for nodes",
		Query:       `resource.type="k8s_node" AND log_id("events")`,
		Category:    "Node",
	},
	{
		Name:        "Kube Proxy Logs",
		Description: "Logs from kube-proxy",
		Query:       `resource.type="k8s_node" AND log_id("kube-proxy")`,
		Category:    "Node",
	},
	{
		Name:        "Container Runtime Logs",
		Description: "Logs from container runtime",
		Query:       `resource.type="k8s_node" AND log_id("container-runtime")`,
		Category:    "Node",
	},
	{
		Name:        "Kubelet Errors",
		Description: "Error or fail logs from kubelet",
		Query:       `resource.type="k8s_node" AND log_id("kubelet") AND jsonPayload.MESSAGE:("error" OR "fail")`,
		Category:    "Node",
	},
	{
		Name:        "Node Logs Collection",
		Description: "Collection of various node-related logs",
		Query:       `resource.type = "k8s_node" logName:( "logs/container-runtime" OR "logs/docker" OR "logs/kube-container-runtime-monitor" OR "logs/kube-logrotate" OR "logs/kube-node-configuration" OR "logs/kube-node-installation" OR "logs/kubelet" OR "logs/kubelet-monitor" OR "logs/node-journal" OR "logs/node-problem-detector")`,
		Category:    "Node",
	},

	// Namespace queries
	{
		Name:        "System Namespaces",
		Description: "Logs from system namespaces",
		Query:       `resource.type = ("k8s_container" OR "k8s_pod") resource.labels.namespace_name = ( "cnrm-system" OR "config-management-system" OR "gatekeeper-system" OR "gke-connect" OR "gke-system" OR "istio-system" OR "knative-serving" OR "monitoring-system" OR "kube-system")`,
		Category:    "Namespace",
	},

	// Container queries
	{
		Name:        "Container Stdout",
		Description: "Stdout logs from containers",
		Query:       `resource.type="k8s_container" AND log_id("stdout")`,
		Category:    "Container",
	},
	{
		Name:        "Container Errors",
		Description: "Stderr error logs from containers",
		Query:       `resource.type="k8s_container" AND log_id("stderr") AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Pod Container Errors",
		Description: "Error logs for a specific pod",
		Query:       `resource.type="k8s_container" AND resource.labels.pod_name="POD_NAME" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Specific Container Errors",
		Description: "Error logs for a specific container in a pod",
		Query:       `resource.type="k8s_container" AND resource.labels.pod_name="POD_NAME" AND resource.labels.container_name="server" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Istio Egress Gateway Errors",
		Description: "Error logs for Istio egress gateway",
		Query:       `resource.type="k8s_container" AND resource.labels.namespace_name="istio-system" AND resource.labels.container_name="egressgateway" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "LoadGenerator Errors",
		Description: "Error logs for loadgenerator app",
		Query:       `resource.type="k8s_container" AND labels."k8s-pod/app"="loadgenerator" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Node Container Errors",
		Description: "Error logs for containers on a specific node",
		Query:       `resource.type="k8s_container" AND labels."compute.googleapis.com/resource_name"=NODE_NAME AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Skaffold Run Errors",
		Description: "Error logs for a specific Skaffold run",
		Query:       `resource.type="k8s_container" AND labels."k8s-pod/app"="loadgenerator" AND labels."k8s-pod/skaffold_dev/run-id"=SKAFFOLD_RUN_ID severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "POST Requests",
		Description: "Logs containing 'POST' in textPayload",
		Query:       `resource.type="k8s_container" AND resource.labels.pod_name="POD_NAME" AND textPayload:"POST" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "GET Requests",
		Description: "Logs with HTTP method GET",
		Query:       `resource.type="k8s_container" AND resource.labels.pod_name="POD_NAME" AND jsonPayload."http.req.method"="GET" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Kube-System Errors",
		Description: "Error logs in kube-system namespace",
		Query:       `resource.type="k8s_container" AND resource.labels.namespace_name="kube-system" AND severity=ERROR`,
		Category:    "Container",
	},
	{
		Name:        "Cloud Error Reporting",
		Description: "Logs for Cloud Error Reporting",
		Query:       `resource.type="k8s_container" AND log_id("clouderrorreporting.googleapis.com/insights")`,
		Category:    "Container",
	},
	{
		Name:        "Specific Container",
		Description: "Logs for a specific container name",
		Query:       `resource.type="k8s_container" AND resource.labels.container_name="CONTAINER_NAME"`,
		Category:    "Container",
	},

	// Control plane queries
	{
		Name:        "API Server Logs",
		Description: "Logs for API server",
		Query:       `resource.type="k8s_control_plane_component" resource.labels.component_name="apiserver" resource.labels.location="CLUSTER_LOCATION" resource.labels.cluster_name="CLUSTER_NAME"`,
		Category:    "Control Plane",
	},
	{
		Name:        "Scheduler Logs",
		Description: "Logs for Scheduler",
		Query:       `resource.type="k8s_control_plane_component" resource.labels.component_name="scheduler" resource.labels.location="CLUSTER_LOCATION" resource.labels.cluster_name="CLUSTER_NAME"`,
		Category:    "Control Plane",
	},
	{
		Name:        "Controller Manager Logs",
		Description: "Logs for Controller Manager",
		Query:       `resource.type="k8s_control_plane_component" resource.labels.component_name="controller-manager" resource.labels.location="CLUSTER_LOCATION" resource.labels.cluster_name="CLUSTER_NAME"`,
		Category:    "Control Plane",
	},
}

type GetSampleQueriesRequest struct {
	Category string `json:"category,omitempty" jsonschema:"Optional category to filter queries by (e.g., 'Cluster', 'Pod', 'Node', 'Container', 'Control Plane', 'Namespace')."`
}

func installGetSampleQueriesTool(s *mcp.Server) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_sample_queries",
		Description: "Get a list of sample LQL queries for common GKE scenarios. Useful for learning how to query logs or finding a starting point for your own queries.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, func(ctx context.Context, _ *mcp.CallToolRequest, req *GetSampleQueriesRequest) (*mcp.CallToolResult, any, error) {
		var filtered []SampleQuery
		if req.Category != "" {
			for _, q := range sampleQueries {
				if q.Category == req.Category {
					filtered = append(filtered, q)
				}
			}
		} else {
			filtered = append(filtered, sampleQueries...)
		}

		b, err := json.MarshalIndent(filtered, "", "  ")
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal sample queries: %w", err)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: string(b),
				},
			},
		}, nil, nil
	})
}
