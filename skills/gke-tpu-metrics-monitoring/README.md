# GKE TPU Metrics Monitoring Skill

This skill provides diagnostics and queries to monitor GKE TPU workloads, nodes, and node pools using GKE system metrics.

## When to Use

Use this skill when:

- You want to monitor the utilization (duty cycle, memory) of GKE TPUs.
- A GKE TPU workload has failed or restarted, and you want to determine if it was due to node preemption, host error, or node pool issues.
- You want to calculate MTTR (Mean Time to Recovery) or MTBI (Mean Time Between Interruptions) for TPU node pools.
- You need to check if TPU nodes are Ready or experiencing Disk/Memory pressure.

## Prerequisites

- GKE version requirements vary by metric (most require 1.27.4+, some 1.28.1+, node status requires 1.32.1+).
- TPU workloads must be configured to export metrics (e.g., `containerPort: 8431` and JAX 0.4.14+).
- GKE System Metrics must be enabled in the cluster.
