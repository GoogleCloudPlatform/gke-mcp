---
name: gke-ai-troubleshooting-handle-disruption-gpu-tpu
description: Diagnose and predict node disruption during Compute Engine host maintenance for GPU and TPU workloads.
---

# Handle Disruption on GPUs and TPUs Troubleshooting

## 🔍 Diagnostic Workflow

### Step 0: Context Acquisition & Execution Trigger

- **Mandatory Context**: `project_id`, `location`, `cluster_name`, `timestamp`.
- **Optional Context**: `node_name`, `workload_name`, `workload_namespace`, `nodepool_name`.
- **CRITICAL EXECUTION DIRECTIVE (STRICTLY ENFORCED)**:
  - **IF ALL 4 MANDATORY PARAMETERS ARE PRESENT**: DO NOT ask the user for confirmation, DO NOT echo parameters back asking for verification, and DO NOT output a text-only plan waiting for user input. YOU MUST IMMEDIATELY INVOKE DIAGNOSTIC TOOLS (`get_k8s_resource`, `query_prometheus`, `query_logs`) IN YOUR VERY FIRST TURN.
  - **IF ANY MANDATORY PARAMETER IS MISSING**: Stop immediately and ask the user to provide ONLY the missing mandatory parameters.

### Step 1: [Low Risk] Check for Upcoming Scheduled Maintenance

- **Action**: Propose running `kubectl` to check if nodes have the scheduled maintenance label indicating an upcoming disruption.
- **Example Command**:
  ```bash
  kubectl get nodes -l cloud.google.com/scheduled-maintenance-time -L cloud.google.com/scheduled-maintenance-time
  ```
- **Interpretation**: The `SCHEDULED-MAINTENANCE-TIME` column shows the Unix epoch time when the VM is scheduled for maintenance. If this label exists, a disruption is guaranteed to occur.

### Step 2: [Low Risk] Investigation via Cloud Monitoring (PromQL)

- **Action**: Call `query_prometheus` tool to query Cloud Monitoring PromQL metrics for the cluster. If `query_prometheus` is unavailable or fails due to missing permissions, provide the PromQL queries below to the user for manual verification in Google Cloud Console.
- **For Past Disruptions**: When investigating past disruptions (e.g., historical timestamps or past interruptions), immediately pass the `timestamp` parameter to `query_prometheus` to check `kubernetes_io:node_interruption_count{monitored_resource="k8s_node", interruption_reason="HW/SW Maintenance"}` around that specific time. If metric value > 0, conclude that Compute Engine host maintenance WAS confirmed as the cause of the past disruption.
- **Example PromQL Queries**:
  ```promql
  # Fetch host maintenance events for nodes
  sum by (interruption_type,interruption_reason)( sum_over_time( kubernetes_io:node_interruption_count{monitored_resource="k8s_node", interruption_reason="HW/SW Maintenance"}[${__interval}]))
  ```
  ```promql
  # See the interruption count aggregated by node pool
  sum by (node_pool_name,interruption_type,interruption_reason)( sum_over_time( kubernetes_io:node_pool_interruption_count{monitored_resource="k8s_node_pool", interruption_reason="HW/SW Maintenance", node_pool_name="<nodepool_name>" }[${__interval}]))
  ```
- **Interpretation**: If `kubernetes_io:node_interruption_count` shows values > 0 for `interruption_reason="HW/SW Maintenance"`, it indicates the underlying Compute Engine VM was interrupted due to scheduled host maintenance.

### Step 3: [Low Risk] Investigation via Cloud Logging & Node Status

- **Action**: Call `query_logs` or `describe_k8s_resource` to filter GKE logs and check node status/taints for active ongoing node maintenance.
- **Active Ongoing Maintenance Detection**:
  - Look for `cloud.google.com/active-node-maintenance` set to `ONGOING` or the `cloud.google.com/impending-node-termination:NoSchedule` taint on the node.
  - **Explicit Conclusion Required**: If `active-node-maintenance` is `ONGOING` or the `impending-node-termination:NoSchedule` taint is present, explicitly inform the user: *"GKE is actively stopping workloads due to ongoing Compute Engine host maintenance."*
  - **Mandatory Taint Warning**: Explicitly advise the user **NOT to tolerate** the `cloud.google.com/impending-node-termination:NoSchedule` taint, as the node will be terminated by GKE regardless.

### Step 4: Conclusion and Resolution

- **Action**: Provide a summary of findings to the user and suggest appropriate mitigation strategies ONLY IF host maintenance events were confirmed or scheduled.
- **Negative Finding Rule**: If NO evidence of scheduled or past host maintenance is found, explicitly conclude that the disruption was NOT caused by Compute Engine host maintenance and report this negative finding. DO NOT recommend configuring graceful termination or opportunistic maintenance when no maintenance events are detected.
- **Reporting Rule**: Signal Only. Report high-signal information indicating that the disruption was caused by Compute Engine host maintenance, specifically affecting the underlying GPU/TPU nodes. DO NOT dump raw logs.
- **Resolutions to Suggest (Only when maintenance is confirmed or scheduled)**:
  1. **Configure Graceful Termination**: Recommend configuring graceful termination by setting `spec.terminationGracePeriodSeconds` (up to 60 minutes) to allow ML workloads (e.g., Orbax checkpointing) to save state upon receiving `SIGTERM` before the node terminates.
  2. **Opportunistic Maintenance**: Recommend configuring Opportunistic Maintenance to trigger host updates automatically when GPU/TPU nodes are idle.
  3. **Capacity Buffer / Resiliency**: Recommend configuring a `PodDisruptionBudget` (PDB) specifying `minAvailable` replicas to maintain availability during disruptions.
