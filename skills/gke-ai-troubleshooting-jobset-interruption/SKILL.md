---
name: gke-ai-troubleshooting-jobset-interruption
description: >
  Systematically diagnose GKE JobSet interruptions, restarts, and preemptions for AI/ML training workloads.
  Identifies preemption events, maintenance interruptions, bad host VMs, unhealthy pods, and coordinator worker failures.
---

# GKE JobSet Interruption Troubleshooting

Use this skill to systematically diagnose and resolve JobSet interruptions, restarts, and preemptions on GKE clusters hosting large-scale AI/ML workloads.

## ⚠️ Prerequisites

- [ ] JobSet metrics package must be enabled in `kube-state-metrics` for your cluster (see [KSM JobSet Metrics](https://docs.cloud.google.com/kubernetes-engine/docs/how-to/kube-state-metrics#ksm-jobset-metrics)).
- [ ] Cloud Logging and Cloud Monitoring must be enabled for the Google Cloud Project.

## 🔍 Diagnostic Workflow

### Step 0: Context Acquisition & Time Window Definition

To begin troubleshooting, acquire the following context from the user:

- **Project ID** (e.g., `customer-ai-project-123`)
- **Cluster Name** (e.g., `tpu-cluster-prod`)
- **Workload Name (JobSet Name)** (e.g., `llama3-70b-training`)
- **Workload Namespace** (e.g., `default`)
- **Issue Time** (e.g., `2026-05-20T08:15:00Z`)

#### Time Handling Rules

1.  **Reject Relative Time**: If the user says "X minutes ago" or "just now", stop and request the exact timestamp or specific time window.
2.  **Window Calculation**: If the user provides a single timestamp `T`, calculate the query window as **`[T - 30m]` to `[T + 30m]`**.
    - Let `Start_Time` = `T - 30m`
    - Let `End_Time` = `T + 30m`

---

### Step 1: Identify JobSet Restarts and Attempts [Low Risk]

Verify if the JobSet is experiencing restart loops and determine the frequency of restarts.

#### Option A: Visual Chart (MQL) - restarts

- **Tool to use**: `monitoring_time_series_chart`
- **MQL Query**:
  ```mql
  fetch prometheus_target
  | metric 'prometheus.googleapis.com/kube_jobset_restarts/gauge'
  | filter resource.cluster_name == '<cluster_name>' && metric.jobset_name == '<workload_name>'
  | align next_older(1m)
  | every 1m
  | group_by [metric.jobset_name], [val: max(value)]
  ```

#### Option B: PromQL Query - restarts

- **Method**: Execute via shell command (requires `BypassSandbox: true`).
- **Command Template**:
  ```bash
  curl -sSf -H "Authorization: Bearer \$(gcloud auth print-access-token)" \
    "https://monitoring.googleapis.com/v1/projects/<project_id>/location/global/prometheus/api/v1/query?query=kube_jobset_restarts%7Bjobset_name%3D%22<workload_name>%22%2Ccluster%3D%22<cluster_name>%22%7D"
  ```
- **Logic**: A non-zero or increasing value for restarts indicates that the JobSet is being actively restarted by the controller due to worker failure or interruption.
- **Automation**: Proceed to Step 2 automatically after reporting findings.

---

### Step 2: Inspect Nodepool Interruptions [Low Risk]

Determine if the JobSet restarts were triggered by physical nodepool-level events (such as spot preemptions, maintenance, or host terminations).

#### A. Metrics Query (Nodepool Interruption Counts)

##### Option A: Visual Chart (MQL) - interruptions

- **Tool to use**: `monitoring_time_series_chart`
- **MQL Query**:
  ```mql
  fetch k8s_node_pool
  | metric 'kubernetes.io/node_pool/interruption_count'
  | filter cluster_name == '<cluster_name>'
  | align next_older(10m)
  | every 10m
  | group_by [metric.interruption_type, metric.interruption_reason, metadata.system.node_pool_name], [val: sum(value)]
  ```

##### Option B: PromQL Query - interruptions

- **Method**: Execute via shell command (requires `BypassSandbox: true`).
- **Command Template**:
  ```bash
  curl -sSf -H "Authorization: Bearer \$(gcloud auth print-access-token)" \
    "https://monitoring.googleapis.com/v1/projects/<project_id>/location/global/prometheus/api/v1/query?query=sum%20by%20%28interruption_type%2C%20interruption_reason%2C%20node_pool_name%2C%20cluster_name%29%20%28avg_over_time%28%7B__name__%3D%22kubernetes.io%2Fnode_pool%2Finterruption_count%22%2C%20monitored_resource%3D%22k8s_node_pool%22%2C%20cluster_name%3D%22<cluster_name>%22%7D%5B10m%5D%29%29"
  ```

#### B. Log Query (Nodepool Life Events)

- **Tool to use**: `query_logs`
- **LQL Log Filter Template**:
  ```sql
  resource.type="gke_nodepool"
  AND resource.labels.cluster_name="<cluster_name>"
  AND timestamp >= "<Start_Time>"
  AND timestamp <= "<End_Time>"
  ```
- **Logic**:
  - **PreemptionEvent**: Spot VMs were preempted, or node was scale-down.
  - **MaintenanceEvent**: Node pool updated or Google scheduled maintenance.
  - **TerminationEvent**: Serious host failures. Check `interruption_reason` or logs for host issues.
- **Automation**: Proceed to Step 3 automatically.

---

### Step 3: Inspect Nodes and Underlying Host VMs [Low Risk]

Correlate node readiness failures with physical host VMs to see if a single faulty host repeatedly fails coordinator pods.

#### A. Metrics Query (Node Ready Status Check)

##### Option A: Visual Chart (MQL) - node status

- **Tool to use**: `monitoring_time_series_chart`
- **MQL Query**:
  ```mql
  fetch k8s_node
  | metric 'kubernetes.io/node/status_condition'
  | filter cluster_name == '<cluster_name>' && metric.condition == 'Ready' && metric.status == 'False'
  | align next_older(1m)
  | every 1m
  | group_by [node_name, metadata.user.gke_nodepool], [val: max(value)]
  ```

##### Option B: PromQL Query - node status

- **Method**: Execute via shell command (requires `BypassSandbox: true`).
- **Command Template**:
  ```bash
  curl -sSf -H "Authorization: Bearer \$(gcloud auth print-access-token)" \
    "https://monitoring.googleapis.com/v1/projects/<project_id>/location/global/prometheus/api/v1/query?query=sum%20by%20%28status%2C%20condition%2C%20node_pool_name%29%20%28%7B__name__%3D%22kubernetes.io%2Fnode%2Fstatus_condition%22%2C%20monitored_resource%3D%22k8s_node%22%2C%20cluster_name%3D%22<cluster_name>%22%2C%20condition%3D%22Ready%22%2C%20status%3D%22False%22%7D%29"
  ```

#### B. Metrics Query (Node-to-Host Metadata Topology Correlation)

- **Tool to use**: `monitoring_time_series_chart`
- **MQL Query**:
  ```mql
  fetch k8s_node
  | metric 'kubernetes.io/node/cpu/total_cores'
  | filter cluster_name == '<cluster_name>'
  | align next_older(1m)
  | every 1m
  | group_by [node_name, metadata.user.gce_topology_host, metadata.user.gke_nodepool], [val: max(value)]
  ```

#### C. Log Query (Node Fault Logs)

- **Tool to use**: `query_logs`
- **LQL Log Filter Template**:
  ```sql
  resource.type="k8s_node"
  AND resource.labels.cluster_name="<cluster_name>"
  AND (textPayload:"host error" OR textPayload:"kernel panic" OR textPayload:"hardware failure" OR textPayload:"NodeNotReady")
  AND timestamp >= "<Start_Time>"
  AND timestamp <= "<End_Time>"
  ```
- **Logic**: Identify if specific nodes are unhealthy (`Ready=False` or `Unknown`) and correlate them to their GCE physical host ID via `metadata.user.gce_topology_host`. Check if the same host is repeatedly failing.
- **Automation**: Proceed to Step 4 automatically.

---

### Step 4: Inspect Pod and Worker / Container Failures [Low Risk]

Analyze pod status phases and retrieve coordinator worker logs to identify application-level crashes or network deadlocks.

#### A. Metrics Query (Pod Lifecycle Phases)

##### Option A: Visual Chart (MQL) - pod phase

- **Tool to use**: `monitoring_time_series_chart`
- **MQL Query**:
  ```mql
  fetch k8s_pod
  | metric 'kubernetes.io/pod/status/phase'
  | filter cluster_name == '<cluster_name>' && pod_name ==~ '<workload_name>.*'
  | align next_older(10m)
  | every 10m
  | group_by [metric.phase], [val: count()]
  ```

##### Option B: PromQL Query - pod phase

- **Method**: Execute via shell command (requires `BypassSandbox: true`).
- **Command Template**:
  ```bash
  curl -sSf -H "Authorization: Bearer \$(gcloud auth print-access-token)" \
    "https://monitoring.googleapis.com/v1/projects/<project_id>/location/global/prometheus/api/v1/query?query=sum%20by%20%28phase%29%20%28avg_over_time%28%7B__name__%3D%22kube_pod_status_phase%22%2C%20cluster%3D%22<cluster_name>%22%2C%20pod%3D~%22<workload_name>.*%22%7D%5B10m%5D%29%29"
  ```

#### B. Metrics Query (Unschedulable Pod Count)

- **Tool to use**: `monitoring_time_series_chart`
- **MQL Query**:
  ```mql
  fetch k8s_pod
  | metric 'kubernetes.io/pod/status/unschedulable'
  | filter cluster_name == '<cluster_name>' && pod_name ==~ '<workload_name>.*'
  | align next_older(10m)
  | every 10m
  | group_by [pod_name], [val: max(value)]
  ```

#### C. Log Query (Worker Container Logs)

- **Tool to use**: `query_logs`
- **LQL Log Filter Template**:
  ```sql
  resource.type="k8s_container"
  AND resource.labels.cluster_name="<cluster_name>"
  AND labels."k8s-pod/jobset_sigs_k8s_io/jobset-name"="<workload_name>"
  AND timestamp >= "<Start_Time>"
  AND timestamp <= "<End_Time>"
  ```
- **Logic**:
  1. Check the pod timeline to spot pending or unschedulable pods.
  2. Use the worker container logs to analyze worker 0 in slice 0 (coordinator) for NCCL timeouts, collective communication issues, or MegaScale hangs.
- **Automation**: Proceed to Resolution.

---

## 🛠️ Resolution Workflow

### Resolution 1: Preemption & Autoscaling Optimizations [Low Risk]

If Step 2 showed high preemption counts on Spot VMs:

- **Action**: Suggest switching critical long-running training workloads to **GKE Reserved/On-Demand VMs** or utilizing **Compact Placement Policies** to minimize defragmentation interruptions.
- **Justification**: Eliminates spot-market preemptions and reduces training restarts.

### Resolution 2: Quarantine Faulty Host VMs [High Risk]

If Step 3 identified a specific host ID (`gce-topology-host`) that consistently fails or triggers restarts across multiple attempts:

- **Action**: Draft a recommendation to cordon/drain the GKE node, delete the underlying GCE VM instance to trigger instance recreation, and open a support ticket with Google Cloud Support specifying the physical host ID.
- **Justification**: GKE auto-repair will recreate the VM instance on healthy physical hardware, preventing infinite restart loops.
- **Automation**: Stop and request explicit user confirmation before cordoning/draining any node.

---

## 📋 Copypaste Checklist

- [ ] Acquire context and compute `[T - 30m, T + 30m]` window.
- [ ] Query JobSet restart attempts.
- [ ] Check Nodepool interruptions (spot preemptions vs. hardware terminations).
- [ ] Query node-to-host mapping and check node logs for physical host errors.
- [ ] Inspect pod timeline status and coordinator worker container logs.
- [ ] Recommend appropriate scheduling strategy (On-demand vs Spot) or host VM quarantining.
