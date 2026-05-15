---
name: gke-workload-scaling
description: Specific workflows for scaling GKE workloads using HPA and VPA, as well as best practices for autoscaling configuration.
---

# GKE Workload Scaling

This skill provides workflows and best practices for scaling applications on Google Kubernetes Engine (GKE). It covers manual scaling, Horizontal Pod Autoscaling (HPA), and Vertical Pod Autoscaling (VPA).

## Workflows

### 1. Manual Scaling

Quickly scale a deployment to a fixed number of replicas. Useful for immediate manual intervention or testing.

**Command:**

```bash
kubectl scale deployment <deployment-name> --replicas=<number> -n <namespace>
```

### 2. Horizontal Pod Autoscaling (HPA)

Automatically scale the number of pods based on observed CPU utilization, memory utilization, or custom metrics.

**Prerequisites:**

- Metrics Server must be running (enabled by default on GKE).
- Containers clearly define resource requests/limits.

**Quick Command:**

```bash
kubectl autoscale deployment <deployment-name> --cpu-percent=50 --min=1 --max=10
```

**Manifest Approach (Recommended):**
Use a YAML manifest for version-controlled configuration.
See [assets/hpa-example.yaml](assets/hpa-example.yaml) for a template.

```bash
kubectl apply -f assets/hpa-example.yaml
```

**Custom Metrics & External Metrics:**
For GKE, the modern and recommended approach for scaling based on Cloud Monitoring metrics (e.g., Pub/Sub queue length) is to use the **External** metric type, which is natively supported by the GKE control plane without requiring the Custom Metrics Adapter. For application-specific metrics exposed via Prometheus, you can use **Google Cloud Managed Service for Prometheus** or the Prometheus Adapter.

### 3. GKE Custom Metrics (`AutoscalingMetric`)

GKE provides a native way to scale workloads based on metrics from Cloud Monitoring or Prometheus metrics exported to Google Cloud Managed Service for Prometheus. This is managed via the GKE-native **`AutoscalingMetric`** Custom Resource Definition (CRD), which avoids the need to deploy and manage the external Custom Metrics Adapter.

**AutoscalingMetric Manifest Approach:**
Define an `AutoscalingMetric` resource to specify which metrics to collect and how to export them.

*Example (Prometheus Metric Filtering by Label):*
To scrape a Prometheus metric `http_requests_total` but filter it to only scale on data points where the label `version` is `v2`:

```yaml
apiVersion: autoscaling.gke.io/v1
kind: AutoscalingMetric
metadata:
  name: http-requests-metric
  namespace: default
spec:
  metric:
    name: http_requests_total
    filter:
      matchLabels:
        version: v2
```

**Renaming Metrics during Export (`exportName`):**
If you need to rename the metric during export (e.g., to prevent naming collisions or simplify references), use the `exportName` parameter:

```yaml
apiVersion: autoscaling.gke.io/v1
kind: AutoscalingMetric
metadata:
  name: web-metric
  namespace: default
spec:
  metric:
    name: original_metric_name
    exportName: exported-name
```

**Referencing Renamed Metrics in HPA:**
When using `AutoscalingMetric` (especially with `exportName`), the HorizontalPodAutoscaler (HPA) must reference the metric using a specific GKE-native prefix format under `spec.metrics[]` (using `type: External`):

Format: `autoscaling.gke.io|<AutoscalingMetric-resource-name>|<export-name>`

*Example HPA configuration for renamed metric:*
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: web-hpa
  namespace: default
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: web-app
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: External
    external:
      metric:
        name: autoscaling.gke.io|web-metric|exported-name
      target:
        type: Value
        value: 100
```

### 4. Vertical Pod Autoscaling (VPA)

Automatically adjust the CPU and memory reservations for your pods to match actual usage. This is critical for right-sizing workloads.

**Prerequisites:**

- VPA must be enabled on the cluster.
  - **Autopilot:** Enabled by default.
  - **Standard:** Must be enabled manually.

**Enable VPA on Standard Cluster:**

```bash
gcloud container clusters update <cluster-name> --enable-vertical-pod-autoscaling --zone <zone>
```

**Update Modes:**

- `Off`: Calculates recommendations but does not apply them. Good for "dry run" analysis.
- `Initial`: Assigns resources only at pod creation time.
- `Auto`: Updates running pods by restarting them if recommendations differ significantly from requests.
- `InPlaceOrRecreate`: Attempts to update Pod resources without recreating the Pod. If in-place update is not possible, it reverts to `Auto` mode (requires GKE 1.34+).

**Example:**
See [assets/vpa-example.yaml](assets/vpa-example.yaml) for a configuration template.

### 4. Cluster Autoscaler

While not a workload-level scaler, the Cluster Autoscaler is essential for ensuring your cluster has enough nodes to run the scaled pods.

**Enable on a Node Pool:**

```bash
gcloud container clusters update <cluster-name> \
    --enable-autoscaling \
    --node-pool <node-pool-name> \
    --min-nodes <min> \
    --max-nodes <max> \
    --zone <zone>
```

## Best Practices

1. **Define Resource Requests:** HPA and VPA rely on accurate resource requests. Always define them in your container specs.
2. **Avoid Metric Conflicts:** Do not configure HPA and VPA to use the same metric (e.g., both CPU). This causes thrashing.
   - _Typical Pattern:_ HPA on CPU, VPA on Memory.
3. **Pod Disruption Budgets (PDBs):** Define PDBs to ensure application availability during scaling events or node upgrades.
4. **HPA Lag:** HPA has a stabilization window (default 5 mins) to prevent rapid fluctuation.
5. **VPA "Auto" Mode Risks:** In "Auto" mode, VPA restarts pods to change resources. Ensure your application handles restarts gracefully (e.g., handles SIGTERM).
   - _Note:_ By default, VPA requires at least 2 replicas to perform evictions. In GKE 1.22+, you can override this by setting `minReplicas` in `PodUpdatePolicy`.
