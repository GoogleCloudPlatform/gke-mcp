# GKE Cost Optimization

This skill provides guidance on optimizing costs for Google Kubernetes Engine (GKE) clusters.

## Overview

Cost optimization in GKE involves tracking costs, setting limits to prevent waste, and rightsizing workloads to match actual usage.

## Workflows

### 1. Enable GKE Cost Allocation & FinOps

GKE cost allocation provides a detailed breakdown of your cluster costs by namespace and labels, enabling effective **FinOps** practices like chargeback and showback.

**Steps:**

1. **Enable Cost Allocation**:
   ```bash
   gcloud container clusters update <cluster-name> \
       --enable-cost-allocation \
       --region <region>
   ```

2. **Enable GKE Usage Metering**: This exports detailed usage data (CPU, memory, storage, network) to BigQuery for advanced analysis.
   ```bash
   gcloud container clusters update <cluster-name> \
       --resource-usage-bigquery-dataset <dataset-name> \
       --region <region>
   ```

**FinOps Lifecycle Integration:**
- **Inform**: Use Cost Allocation and BigQuery exports to gain visibility into spend.
- **Optimize**: Apply Rightsizing, Spot Pods, and Compute Classes based on visibility data.
- **Operate**: Continuously monitor and adjust budgets/quotas to align with business value.

### 2. Configure Resource Quotas

Resource quotas restrict the total resource consumption in a namespace, preventing any single tenant from consuming all cluster resources.

**Example ResourceQuota Manifest:**

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: compute-quota
  namespace: my-namespace
spec:
  hard:
    requests.cpu: "4"
    requests.memory: 16Gi
    limits.cpu: "8"
    limits.memory: 32Gi
```

### 3. Rightsizing Strategies

Rightsizing involves adjusting the requested resources of your workloads to match their actual utilization.

- **Use VPA in Recommender Mode**: Let VPA observe usage and recommend CPU and memory requests.
- **Autopilot Rightsizer**: For Autopilot clusters, GKE provides an automated rightsizer that can adjust resource limits to match actual usage over time.
- **Review Cost Recommendations**: Check the Google Cloud Console for GKE cost optimization recommendations.

### 4. Optimize with Spot Pods (Autopilot)

In Autopilot, you can use Spot Pods to run fault-tolerant workloads at a significant discount.

**Example Pod Spec for Spot:**

```yaml
spec:
  terminationGracePeriodSeconds: 25
  containers:
  - name: my-app
    image: my-image
  nodeSelector:
    cloud.google.com/gke-spot: "true"
```

### 5. Use Compute Classes (Autopilot)

GKE Autopilot allows you to select different hardware configurations (Compute Classes) for your pods based on their needs (e.g., `performance`, `scale-out`, `gpu`).

**Example Pod Spec with Compute Class:**

```yaml
spec:
  nodeSelector:
    cloud.google.com/compute-class: "performance"
  containers:
  - name: my-app
    resources:
      requests:
        cpu: "2"
        memory: "4Gi"
```

## Best Practices

1. **Autopilot for Most Workloads**: Use Autopilot to automatically benefit from node rightsizing and pay-per-pod pricing.
2. **Enable Cost Allocation**: Always enable GKE cost allocation to understand where your money is going.
3. **Use Resource Quotas**: Enforce resource quotas in multi-tenant clusters to prevent cost runaways.
4. **Leverage Spot VMs/Pods**: Use Spot for fault-tolerant, stateless workloads to save up to 91%.
5. **Automate Scaling**: Use Cluster Autoscaler (Standard) or native Autopilot scaling along with HPA/VPA to ensure you only pay for what you need.
