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

package k8s

import (
	"context"
	"fmt"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/tools/params"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type getK8SRolloutStatusArgs struct {
	params.Cluster
	ResourceType string `json:"resourceType" jsonschema:"Required. The type of resource to check. e.g. \"deployment\", \"daemonset\", \"statefulset\"."`
	Name         string `json:"name" jsonschema:"Required. The name of the resource to check."`
	Namespace    string `json:"namespace" jsonschema:"Required. The namespace of the resource."`
}

func (h *handlers) getK8SRolloutStatus(ctx context.Context, _ *mcp.CallToolRequest, args *getK8SRolloutStatusArgs) (*mcp.CallToolResult, any, error) {
	if args == nil {
		return params.ErrorResult(fmt.Errorf("args cannot be nil")), nil, nil
	}
	if args.Namespace == "" {
		return params.ErrorResult(fmt.Errorf("namespace is required")), nil, nil
	}
	clusterPath := args.ClusterPath()

	discoveryClient, err := h.provider.DiscoveryClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get discovery client: %w", err)), nil, nil
	}

	_, gvk, _, err := ResolveGVR(ctx, discoveryClient, args.ResourceType)
	if err != nil {
		return params.ErrorResult(err), nil, nil
	}

	namespace := args.Namespace

	clientset, err := h.provider.KubernetesClient(ctx, clusterPath)
	if err != nil {
		return params.ErrorResult(fmt.Errorf("failed to get kubernetes client: %w", err)), nil, nil
	}

	var msg string
	switch gvk.Kind {
	case "Deployment":
		dep, err := clientset.AppsV1().Deployments(namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get deployment: %w", err)), nil, nil
		}
		msg = checkDeploymentRolloutStatus(dep)
	case "DaemonSet":
		ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get daemonset: %w", err)), nil, nil
		}
		msg = checkDaemonSetRolloutStatus(ds)
	case "StatefulSet":
		ss, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if err != nil {
			return params.ErrorResult(fmt.Errorf("failed to get statefulset: %w", err)), nil, nil
		}
		msg = checkStatefulSetRolloutStatus(ss)
	default:
		return params.ErrorResult(fmt.Errorf("rollout status not supported for resource of kind %q", gvk.Kind)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: msg},
		},
	}, nil, nil
}

func checkDeploymentRolloutStatus(deployment *appsv1.Deployment) string {
	specReplicas := int32(1)
	if deployment.Spec.Replicas != nil {
		specReplicas = *deployment.Spec.Replicas
	}

	if deployment.Generation > deployment.Status.ObservedGeneration {
		return fmt.Sprintf("deployment %q generation %d is still being processed", deployment.Name, deployment.Generation)
	}
	if deployment.Status.UpdatedReplicas < specReplicas {
		return fmt.Sprintf("waiting for rollout to finish: %d out of %d new replicas have been updated", deployment.Status.UpdatedReplicas, specReplicas)
	}
	if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
		return fmt.Sprintf("waiting for rollout to finish: %d old replicas are pending termination", deployment.Status.Replicas-deployment.Status.UpdatedReplicas)
	}
	if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
		return fmt.Sprintf("waiting for rollout to finish: %d of %d updated replicas are available", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
	}
	return fmt.Sprintf("deployment %q successfully rolled out", deployment.Name)
}

func checkDaemonSetRolloutStatus(daemonSet *appsv1.DaemonSet) string {
	if daemonSet.Generation > daemonSet.Status.ObservedGeneration {
		return fmt.Sprintf("daemonset %q generation %d is still being processed", daemonSet.Name, daemonSet.Generation)
	}
	if daemonSet.Status.UpdatedNumberScheduled < daemonSet.Status.DesiredNumberScheduled {
		return fmt.Sprintf("waiting for rollout to finish: %d out of %d new pods have been updated", daemonSet.Status.UpdatedNumberScheduled, daemonSet.Status.DesiredNumberScheduled)
	}
	if daemonSet.Status.NumberAvailable < daemonSet.Status.DesiredNumberScheduled {
		return fmt.Sprintf("waiting for rollout to finish: %d of %d updated pods are available", daemonSet.Status.NumberAvailable, daemonSet.Status.DesiredNumberScheduled)
	}
	if daemonSet.Status.NumberMisscheduled > 0 {
		return fmt.Sprintf("waiting for rollout to finish: %d pods are misscheduled", daemonSet.Status.NumberMisscheduled)
	}
	return fmt.Sprintf("daemonset %q successfully rolled out", daemonSet.Name)
}

func checkStatefulSetRolloutStatus(statefulSet *appsv1.StatefulSet) string {
	if statefulSet.Spec.UpdateStrategy.Type == appsv1.OnDeleteStatefulSetStrategyType {
		return fmt.Sprintf("statefulset %q configured with OnDelete update strategy, rollout is considered complete", statefulSet.Name)
	}

	if statefulSet.Generation > statefulSet.Status.ObservedGeneration {
		return fmt.Sprintf("statefulset %q generation %d is still being processed", statefulSet.Name, statefulSet.Generation)
	}

	replicas := int32(1)
	if statefulSet.Spec.Replicas != nil {
		replicas = *statefulSet.Spec.Replicas
	}

	if statefulSet.Status.ReadyReplicas < replicas {
		return fmt.Sprintf("waiting for rollout to finish: %d of %d pods are ready", statefulSet.Status.ReadyReplicas, replicas)
	}

	if statefulSet.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType {
		if statefulSet.Spec.UpdateStrategy.RollingUpdate != nil &&
			statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			// Partitioned update.
			partition := *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition
			if statefulSet.Status.UpdatedReplicas < replicas-partition {
				return fmt.Sprintf("waiting for partitioned rollout to finish: %d of %d pods in partition have been updated", statefulSet.Status.UpdatedReplicas, replicas-partition)
			}
			return fmt.Sprintf("statefulset %q successfully rolled out", statefulSet.Name)
		}
		// Full rolling update.
		if statefulSet.Status.UpdatedReplicas < replicas {
			return fmt.Sprintf("waiting for rollout to finish: %d of %d pods have been updated", statefulSet.Status.UpdatedReplicas, replicas)
		}
	}

	if statefulSet.Status.UpdateRevision != "" && statefulSet.Status.CurrentRevision != statefulSet.Status.UpdateRevision {
		return fmt.Sprintf("waiting for statefulset rolling update to complete: pods are being updated to revision %s", statefulSet.Status.UpdateRevision)
	}

	return fmt.Sprintf("statefulset %q successfully rolled out", statefulSet.Name)
}
