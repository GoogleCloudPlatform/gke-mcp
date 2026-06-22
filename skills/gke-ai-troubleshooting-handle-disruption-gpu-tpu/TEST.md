# Test Plan

## Prerequisites

- A GKE cluster with a GPU or TPU node pool.
- `gcloud` CLI authenticated.

## Test Cases

### 1. Identify Upcoming Maintenance

1. Identify a node with an upcoming scheduled maintenance (if available) or simulate the label:
   `kubectl label node <node-name> cloud.google.com/scheduled-maintenance-time=1733083200`
2. Ask the agent: "Can you check if my GPU nodes have upcoming maintenance?"
3. Verify the agent runs the `kubectl` check and identifies the timestamp.

### 2. Diagnose Active Maintenance via Logs

1. Trigger a simulated host maintenance event on a GPU node pool using:
   `gcloud beta compute instances simulate-maintenance-event <instance-name> --zone=<zone>`
2. Ask the agent: "My GPU workloads just died, is it due to host maintenance?"
3. Verify the agent queries Cloud Logging and identifies the `cloud.google.com/active-node-maintenance` log and `impending-node-termination:NoSchedule` taint.

### 3. Verify Remediation Advice

1. Ask the agent: "How do I prevent my training jobs from crashing during maintenance?"
2. Verify the agent suggests configuring `spec.terminationGracePeriodSeconds` and setting up Opportunistic Maintenance.
