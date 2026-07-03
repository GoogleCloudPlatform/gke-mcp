---
name: verify_unused
description: Verifies if a GKE or Kubernetes cluster is unused (no active compute, external exposure, or persistent data) before allowing deletion. Evaluates external exposure (LoadBalancer Service, Ingress, Gateway, MultiClusterIngress), persistent data (Bound PVC), and active compute (Running/Pending Pods in user namespaces) with low-overhead queries and fail-close timeouts.
---

# verify_unused

Use this skill to safety-check whether a GKE or Kubernetes cluster is active or unused **before** executing any cluster deletion commands (`DeleteCluster`, `gcloud container clusters delete`, or `kubectl delete`). This prevents accidental data loss, service interruption, and destruction of active environments.

## Critical Rule
> [!IMPORTANT]
> **NEVER** delete a GKE or Kubernetes cluster without running this verification skill first or verifying that `deletion_policy: VERIFY_UNUSED` is enforced. If this check returns `[ACTIVE]` (exit code `1`) or fails due to network/server timeout (`[FAIL-CLOSE]`, exit code `2`), you must **block the deletion** and prompt the user for explicit confirmation or intervention.

## Safety Criteria

The safety verification evaluates three core heuristics. If any condition is met, deletion is blocked:

### 1. External Exposure
* **Failure Condition:** Presence of $\ge 1$ `Service` of type `LoadBalancer`, OR any `Ingress`, `Gateway` (Gateway API), or `MultiClusterIngress` resource in any namespace.
* **Rationale:** These resources indicate the cluster is actively serving (or configured to serve) ingress traffic. Deleting the cluster would cause an immediate service disruption for users or connected systems.

### 2. Persistent Data
* **Failure Condition:** Presence of $\ge 1$ `PersistentVolumeClaim` (PVC) in the `Bound` state in any namespace.
* **Rationale:** A Bound PVC indicates a persistent disk attached containing application state. Deleting the cluster could result in permanent, unrecoverable data loss.

### 3. Active Compute
* **Failure Condition:** Presence of $\ge 1$ `Pod` in `Running` or `Pending` state in any user namespace (includes `default`, but excludes system namespaces like `kube-system` or those prefixed with `gke-`).
* **Rationale:** Running pods represent active compute, and Pending pods indicate a system attempting to schedule or scale work. System namespaces (`kube-system`, `gke-*`) are excluded to avoid blocking deletion on cluster control-plane and addon overhead.

## Performance Overhead & Reliability

To ensure checks add minimal overhead and operate safely under adverse network conditions:
* **Low-Overhead Queries (Early Return):** Because the policy only needs to confirm the existence of at least one active resource to block deletion, lists use standard chunk sizes (`--chunk-size=500`) and the Python script terminates immediately upon finding the first offending resource.
  Note that passing `--chunk-size=1` with `kubectl -o json` is avoided because `kubectl` sequentially loops through all pages before formatting output, which causes N sequential API requests.
* **Bounded Timeouts & Fail-Close Behavior:** Synchronous API queries enforce explicit timeouts (`--request-timeout`).
  If the checks do not complete within the timeout window or encounter API errors, the operation **fails close** (exit code `2`), blocking cluster deletion rather than assuming safety during control-plane latency or outages.

## Usage Instructions

### Step 1: Execute Verification Script
Execute the bundled safety verification script (`scripts/verify_unused.py` located inside this skill folder) against the target cluster:

```bash
python3 <skill_directory>/scripts/verify_unused.py --timeout 5.0
```

*(Optional flags: `--kubeconfig <path>` or `--context <name>` can be passed to target specific environments).*

### Step 2: Interpret Output & Exit Codes

* **Exit Code `0` (`[UNUSED]`)**:
  ```text
  [UNUSED] Cluster is verified unused (no active compute, exposure, or persistent data).
  It is safe to proceed with cluster deletion.
  ```
  *Action:* Safe to proceed with cluster deletion.

* **Exit Code `1` (`[ACTIVE]`)**:
  ```text
  [ACTIVE] Cluster is currently in use! Deletion blocked.
  Active workloads/resources detected:
    - External Exposure: Service prod/frontend is of type LoadBalancer
  ```
  *Action:* **DO NOT delete the cluster.** Stop immediately and present the detected active resources to the user.

* **Exit Code `2` (`[FAIL-CLOSE]`)**:
  ```text
  [FAIL-CLOSE] Cluster safety check failed due to query error or timeout!
  ```
  *Action:* **DO NOT delete the cluster.** Stop and report that verification failed close due to query error or API server timeout.
