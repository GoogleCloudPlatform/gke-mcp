---
name: gke-ai-troubleshooting-tpu-connection-failure-vbar-oom
description: >
  Diagnose and prevent `vbar_control_agent` segfaults and OOMs caused by race
  conditions during TPU device resets and frequent metrics collection (e.g.
  every 3s). Use when TPU slice initialization fails or `vbar_control_agent`
  crashes on TPU v6e nodes.
---

# TPU Connection Failure and VBAR OOM Troubleshooting

Use this skill to systematically diagnose and prevent `vbar_control_agent`
segfaults and Out-Of-Memory (OOM) errors on TPU v6e nodes.

## ⚠️ Prerequisites

- [ ] Cloud Logging must be enabled for the project.
- [ ] Access to the project and cluster via `gcloud` or equivalent tool.

## 🔍 Diagnostic Workflow

### Step 0: Context Acquisition & Time Window Definition

To begin troubleshooting, acquire the following context from the user:

- **Project ID** (e.g., `customer-ai-project-123`)
- **Cluster Name** (e.g., `tpu-cluster-prod`)
- **Node Name or Instance ID** (e.g., `tpu-node-1`)
- **Workload Name (JobSet Name)** (e.g., `my-training-job-456`)
- **Workload Namespace**
- **Issue Time** (e.g., `2026-04-14T20:00:00Z`)

#### Time Handling Rules

1.  **Reject Relative Time**: If the user says "X minutes ago" or "just now",
    stop and ask for the exact timestamp or a specific time window.
2.  **Window Calculation**: If the user provides a start time or an "around"
    time `T`, calculate the query window as **`[T - 30m]` to `[T + 30m]`**.
    - Let `Start_Time` = `T - 30m`
    - Let `End_Time` = `T + 30m`

### Step 1: Check for `vbar_control_agent` OOMs

Look for specific `out of memory` messages from `vbar_control_agent` in serial
console logs.

- **Tool to use**: `query_logs`
- **Filter Templates**:

**Serial Console Logs (OOMs):**

```sql
logName="projects/<project_id>/logs/serialconsole.googleapis.com%2fserial_port_1_output"
AND labels."compute.googleapis.com/resource_name"="<node_name>"
AND SEARCH(text_payload, "Memory cgroup out of memory: Killed process .* (vbar_control_ag)")
AND timestamp >= "<Start_Time>"
AND timestamp <= "<End_Time>"
```

- **Logic**: Presence of `Memory cgroup out of memory` messages related to
  `vbar_control_agent`. Stack traces pointing to
  `libtpu::tpunetd::VBARControlHelper::MetricsReadFromVBAR` are a strong indicator.
- **Automation**: Proceed to next step automatically after reporting findings.
- **Reference**: See `references/failure_signatures.md` for example log patterns.

### Step 2: Investigate `tpu-device-plugin` Metrics Fetch Failures [Low Risk]

Check if `tpu-device-plugin` is reporting metric fetch failures.

- **Tool to use**: `query_logs`
- **Filter Template**:

```sql
resource.type="k8s_container"
AND resource.labels.project_id="<project_id>"
AND resource.labels.cluster_name="<cluster_name>"
AND resource.labels.container_name="tpu-device-plugin"
AND severity=ERROR
AND textPayload:"metrics fetch failed for .* deviceID and .* device path with error: checksum didn't match with the metrics data. Corrupt data found"
AND timestamp >= "<Start_Time>"
AND timestamp <= "<End_Time>"
```

- **Logic**: Errors indicating "metrics fetch failed" with "checksum didn't
  match" suggest vBAR memory corruption.
- **Automation**: Proceed to next step automatically after reporting findings.

### Step 3: Check for Custom Metrics Collection Usage [Low Risk]

Inquire with the user about any custom TPU metrics collection mechanisms they
have deployed.

- **Action**: Ask the user if they are using custom scripts or agents (e.g.,
  using `libtpu.sdk.tpumonitoring`) that frequently query `GetHostMetrics` from
  `vBAR Control Agent`.
- **Logic**: Confirmation of custom metrics collection helps confirm the race
  condition hypothesis.
- **Automation**: Stop and wait for user response before proceeding to resolution.

## 🛠️ Resolution Workflow

### Resolution 1: Temporarily Disable Custom Metrics Collection [High Risk]

If a custom metrics collection agent is identified, advise the user to
temporarily disable it.

- **Action**: Recommend disabling the custom metrics collector.
- **Justification**: Prevents reads from vBAR during device resets, stopping
  crashes and OOMs.
- **Automation**: Stop and request explicit user approval in the bug thread
  before making this recommendation or taking action.

### Resolution 2: Await `vbar_control_agent` Resiliency Update [Low Risk]

Advise the user that a permanent fix will be available in a future GKE version.

- **Action**: Recommend upgrading GKE when the fix is available.
- **Justification**: The updated agent will be resilient to memory corruption
  and gracefully handle reads from unbound vBARs.
- **Automation**: Proceed to report this finding.

## 📋 copypaste checklist

- [ ] Acquire context and compute `[T - 30m, T + 30m]` window.
- [ ] Check for `vbar_control_agent` segfaults and OOMs using `query_logs`.
- [ ] Investigate `tpu-device-plugin` failures using `query_logs`.
- [ ] Ask user about custom metrics collection usage.
- [ ] Advise disabling custom metrics collection (High Risk) if applicable.
- [ ] Advise awaiting resiliency update.
