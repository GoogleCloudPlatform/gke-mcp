---
name: gke-ai-troubleshooting-handle-disruption-gpu-tpu
description: Diagnose and predict node disruption during Compute Engine host maintenance for GPU and TPU workloads.
---

# Handle Disruption on GPUs and TPUs Troubleshooting

## 🔍 Diagnostic Workflow

### Step 0: Context Acquisition

- **Mandatory**: <project_id>, <location>, <cluster_name>, <timestamp>.
- **Optional**: <node_name>, <workload_name>, <workload_namespace>, <nodepool_name>.

### Step 1: [Low Risk] Check for Upcoming Scheduled Maintenance

- **Action**: Propose running `kubectl` to check if nodes have the scheduled maintenance label indicating an upcoming disruption.
- **Example Command**:
  ```bash
  kubectl get nodes -l cloud.google.com/scheduled-maintenance-time -L cloud.google.com/scheduled-maintenance-time
  ```
- **Interpretation**: The `SCHEDULED-MAINTENANCE-TIME` column shows the Unix epoch time when the VM is scheduled for maintenance. If this label exists, a disruption is guaranteed to occur.

### Step 2: [Low Risk] Investigation via Cloud Monitoring (PromQL)

- **Action**: Call any available monitoring tool or provide PromQL for manual verification.
- **Example Query**:
  ```promql
  # Fetch host maintenance events for nodes
  sum by (interruption_type,interruption_reason)( sum_over_time( kubernetes_io:node_interruption_count{monitored_resource="k8s_node", interruption_reason="HW/SW Maintenance"}[${__interval}]))
  ```
  ```promql
  # See the interruption count aggregated by node pool
  sum by (node_pool_name,interruption_type,interruption_reason)( sum_over_time( kubernetes_io:node_pool_interruption_count{monitored_resource="k8s_node_pool", interruption_reason="HW/SW Maintenance", node_pool_name="<nodepool_name>" }[${__interval}]))
  ```
- **Interpretation**: If `kubernetes_io:node_interruption_count` shows values > 0 for `interruption_reason="HW/SW Maintenance"`, it indicates the underlying Compute Engine VM was interrupted due to scheduled host maintenance.

### Step 3: [Low Risk] Investigation via Cloud Logging

- **Action**: Call `query_logs` or instruct the user to filter their GKE logs for graceful termination events.
- **Guidance**: Look for instances where `cloud.google.com/active-node-maintenance` is set to `ONGOING`. You can also check if GKE has applied the `cloud.google.com/impending-node-termination:NoSchedule` taint to restrict new workloads from being scheduled.
- **Interpretation**:
  - `cloud.google.com/active-node-maintenance` set to `ONGOING` means workloads are actively being stopped by GKE due to host maintenance.
  - `cloud.google.com/impending-node-termination:NoSchedule` taint means GKE has cordoned the node to prevent new Pods from being scheduled on the terminating node. DO NOT recommend tolerating this taint.

### Step 4: Conclusion and Resolution

- **Action**: Provide a summary of findings to the user and suggest appropriate mitigation strategies if host maintenance events were confirmed or scheduled.
- **Reporting Rule**: Signal Only. Report high-signal information indicating that the disruption was caused by Compute Engine host maintenance, specifically affecting the underlying GPU/TPU nodes. DO NOT dump raw logs.
- **Resolutions to Suggest**:
  1. **Configure Graceful Termination**: For workloads that need time to save state (e.g., ML frameworks checkpointing via Orbax), follow the guide to [Enable disruption handling](https://docs.cloud.google.com/kubernetes-engine/docs/concepts/handle-disruption-gpu-tpu#enabling-handling) and set `spec.terminationGracePeriodSeconds` to handle the `SIGTERM` signal.
  2. **Opportunistic Maintenance**: To automatically trigger maintenance when GKE detects that GPU/TPU nodes are idle, configure [Opportunistic Maintenance](https://docs.cloud.google.com/kubernetes-engine/docs/concepts/handle-disruption-gpu-tpu#opportunistic-maintenance).
  3. **Capacity Buffer / Resiliency**: Ensure your workload uses a `PodDisruptionBudget` to maintain `minAvailable` replicas during disruptions.
