---
name: gke-tpu-dynamic-slices-monitoring
description: >
  Monitor and manage GKE TPU Dynamic Slices custom resources. Use when
  checking slice lifecycle states, troubleshooting failed slice creations (e.g.
  SliceCreationFailed, FAILED), running single or multi-slice workloads, or
  safely deleting/disabling slices.
---

# GKE TPU Dynamic Slices Monitoring & Management

Use this skill to monitor the status of TPU Slice custom resources, troubleshoot provisioning failures, deploy workloads on dynamic slices, and perform cleanups.

## ⚠️ Prerequisites

- [ ] Cloud Logging must be enabled for the project.
- [ ] `kubectl` and `gcloud` CLIs must be configured to access the GKE cluster.

## 🔍 Diagnostic Workflow

### Step 0: Context Acquisition & Time Window Definition

To begin monitoring or troubleshooting, acquire the following context from the user:

- **Project ID** (e.g., `my-gcp-project`)
- **Cluster Name** (e.g., `tpu-cluster`)
- **Region/Zone** (e.g., `us-central1-a`)
- **Slice Name** (e.g., `test-slice`)
- **Issue Time** (Optional, only needed if diagnosing a historical failure; e.g., `2026-06-25T12:00:00Z`)

#### Time Handling Rules (If Issue Time is provided)

1.  **Reject Relative Time**: If the user says "a few minutes ago" or "just now", stop and ask for the exact timestamp.
2.  **Window Calculation**: Calculate the query window as **`[T - 30m]` to `[T + 30m]`**.
    - Let `Start_Time` = `T - 30m`
    - Let `End_Time` = `T + 30m`

---

### Step 1: Describe the Slice Custom Resource [Low Risk]

Check the current state, topology, and conditions of the specified Slice custom resource.

- **Tool to use**: `run_command`
- **Command**:
  ```bash
  kubectl describe slice <Slice Name>
  ```

#### State & Reason Analysis

Analyze the `Status.Conditions` (especially `Type: Ready` and its `Reason` and `Status`):

| Lifecycle State / Reason  | Meaning                                                                                                                                                              | Recommended Action                                                                                                                             |
| :------------------------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------- |
| **`SliceNotCreated`**     | GKE Slice Controller is initializing the slice and performing resource checks.                                                                                       | **[Auto]** Wait a few minutes and run the command again.                                                                                       |
| **`SliceCreationFailed`** | Prerequisites validation failed (e.g., selected nodes don't exist, nodes are already used by another slice, or the topology doesn't match the number of partitions). | **[Manual]** Verify that the selected nodes exist, are not already in use by another slice, and that the topology matches the partition count. |
| **`ACTIVATING`**          | GKE is actively forming and provisioning the TPU slice.                                                                                                              | **[Auto]** Wait and monitor node provisioning.                                                                                                 |
| **`ACTIVE`**              | The TPU slice is successfully formed and ready to host workloads.                                                                                                    | **[Auto]** Proceed to deploy or check workloads.                                                                                               |
| **`ACTIVE_DEGRADED`**     | The slice is usable, but one or more sub-blocks are degraded.                                                                                                        | **[Manual]** Monitor workload logs closely for interconnect or device errors. Check for faulty node/host VMs.                                  |
| **`FAILED`**              | GKE failed to form the TPU slice (e.g., selected nodes are not part of the same reservation block, or are not in the same reservation).                              | **[Manual]** Ensure all selected nodes belong to the same reservation and reservation block.                                                   |
| **`DEACTIVATING`**        | The slice is dismantling (either triggered by user deletion or a critical systemic failure).                                                                         | **[Auto]** Wait for dismantling to finish, or patch finalizers if it gets stuck (see Resolution 1).                                            |
| **`INCOMPLETE`**          | The terminal phase before the Slice CR is deleted from the cluster.                                                                                                  | **[Auto]** No action required; the resource will be removed shortly.                                                                           |

---

### Step 2: Verify Workload Specification [Low Risk]

Ensure the user's workload manifests are configured correctly to target the dynamic slice.

#### 1. Single-Slice Workload Requirements

Check that the Pod template contains the following annotations and selectors:

- **Annotations**:
  - `cloud.google.com/gke-tpu-slice-topology: "<Topology>"` (e.g., `"4x4x4"`)
- **NodeSelector**:
  - `cloud.google.com/gke-tpu-topology: "<Topology>"` (e.g., `"4x4x4"`)
  - `cloud.google.com/gke-tpu-accelerator: "<Accelerator Type>"` (e.g., `"tpu7x"`)
  - `cloud.google.com/gke-tpu-slice: "<Slice Name>"` (e.g., `"test-slice"`)

#### 2. Multi-Slice (JobSet) Workload Requirements

If deploying a multi-slice JobSet, verify:

- **JobSet Annotation**:
  - `alpha.jobset.sigs.k8s.io/exclusive-topology: cloud.google.com/gke-tpu-slice`
- **Pod Template Annotations**:
  - `cloud.google.com/gke-tpu-slice-topology: "<Topology>"`
- **Pod Template NodeSelector**:
  - `cloud.google.com/gke-tpu-topology: "<Topology>"`
  - `cloud.google.com/gke-tpu-accelerator: "<Accelerator Type>"`
  - _Note: Do NOT manually specify `cloud.google.com/gke-tpu-slice` in the nodeSelector; JobSet handles slice assignment automatically._

---

## 🛠️ Resolution & Management Workflow

### Resolution 1: Force Delete a Stuck Slice [High Risk]

If a slice is stuck in `DEACTIVATING` or deletion hangs indefinitely (often due to finalizers awaiting resource cleanup), patch the resource to remove finalizers.

- **Tool to use**: `run_command`
- **Command**:
  ```bash
  kubectl patch slice <Slice Name> --type json -p='[{"op": "remove", "path": "/metadata/finalizers"}]'
  ```
- **Automation**: Stop and ask for explicit user confirmation before executing this patch.

---

### Resolution 2: Disable and Clean Up Slice Controller [High Risk]

If the user needs to completely disable dynamic slicing and the GKE Slice Controller:

1.  **Check for existing Slices**:
    ```bash
    kubectl get slice -A
    ```
    _Ensure all slices are deleted before disabling the controller._
2.  **Disable Slice Controller via gcloud**:
    ```bash
    gcloud container clusters update <Cluster Name> \
        --location=<Region/Zone> \
        --no-enable-slice-controller
    ```
3.  **Delete the Slice CRD**:
    ```bash
    kubectl delete crd slices.accelerator.gke.io
    ```
4.  **Clean up Node Labels**:
    Remove GKE TPU Slice labels from all nodes in the cluster:
    ```bash
    kubectl label nodes --all cloud.google.com/gke-tpu-slice- cloud.google.com/gke-tpu-slice-topology-
    ```

- **Automation**: Stop and ask for explicit user confirmation before executing any disabling or destructive cleanup commands.

---

## 📋 Copypaste Checklist

- [ ] Acquire project context and slice names.
- [ ] Inspect slice status via `kubectl describe slice`.
- [ ] Inspect Kubernetes events and GKE control plane logs for errors.
- [ ] Validate workload manifest annotations and nodeSelectors.
- [ ] If deletion hangs, patch finalizers (requires user approval).
- [ ] If decommissioning, follow the full Slice Controller disabling sequence.
