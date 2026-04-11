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

### 3. Vertical Pod Autoscaling (VPA)

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
- `InPlaceOrRecreate`: Attempts to update Pod resources without recreating the Pod using **In-Place Pod Resizing** (requires GKE 1.34+).

### 4. Multi-dimensional Pod Autoscaling (MPA)

For workloads that need to scale both horizontally (more pods) and vertically (larger pods), use MPA. This is especially useful for workloads where CPU and memory usage are not perfectly correlated.

### 5. Cluster Autoscaler

While not a workload-level scaler, the Cluster Autoscaler is essential for ensuring your cluster has enough nodes to run the scaled pods. In **Autopilot**, this is handled automatically.

### 6. Karpenter for GKE (Standard Clusters)

For high-velocity, multi-dimensional scaling in Standard clusters, use **Karpenter**. Karpenter observes unschedulable pods and quickly provisions the right-sized nodes for them.

**Advantages:**
- Faster node provisioning compared to Cluster Autoscaler.
- Directly creates Compute Engine instances without node pool overhead.
- Intelligent workload consolidation to reduce costs.

### 7. Image Streaming for Faster Startup

Reduce the time it takes for new pods to start (Cold Start) by streaming image data. This is particularly effective for large images (>1GB).

**Usage:**
- **Autopilot**: Enabled by default for images on AR/GCR.
- **Standard**: Enable on the node pool.
  ```bash
  gcloud container node-pools update <pool-name> --cluster <cluster-name> --enable-image-streaming
  ```

## Best Practices

1. **Prefer Autopilot**: Use Autopilot to benefit from automatic node provisioning and built-in scaling optimizations.
2. **Define Resource Requests**: HPA and VPA rely on accurate resource requests. Always define them in your container specs.
3. **Avoid Metric Conflicts**: Do not configure HPA and VPA to use the same metric (e.g., both CPU). This causes thrashing.
   - _Typical Pattern:_ HPA on CPU, VPA on Memory.
4. **Pod Disruption Budgets (PDBs)**: Define PDBs to ensure application availability during scaling events or node upgrades.
5. **HPA Lag**: HPA has a stabilization window (default 5 mins) to prevent rapid fluctuation.
6. **VPA "Auto" Mode Risks**: In "Auto" mode, VPA restarts pods to change resources. Ensure your application handles restarts gracefully (e.g., handles SIGTERM).
   - _Note:_ By default, VPA requires at least 2 replicas to perform evictions. In GKE 1.22+, you can override this by setting `minReplicas` in `PodUpdatePolicy`.
