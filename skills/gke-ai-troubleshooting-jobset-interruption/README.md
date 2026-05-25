# Skill: GKE AI Troubleshooting - JobSet Interruption

This skill provides an automated diagnostic and resolution workflow for GKE JobSet interruptions, restarts, and preemption events on large-scale AI/ML training workloads.

## What is this issue

Large-scale AI/ML training jobs (e.g., LLMs, reinforcement learning) orchestrated using `JobSet` on GKE often experience interruptions. These interruptions can stem from:

1.  **Preemption / Maintenance Events**: Underneath nodes (e.g., Spot VMs) being preempted or undergoing regular GKE maintenance.
2.  **Host / Hardware Failures**: Severe physical node failures or hypervisor/hardware issues.
3.  **Worker/Container Failures**: Application-level crashes, deadlocks, NCCL timeouts, or collective communication failures.

In large slices, a single worker failing or a single physical host malfunctioning can repeatedly crash coordinator pods, leading to an infinite cycle of training restarts.

## When to use this skill

- A GKE JobSet shows frequent, unexpected restarts or failure attempts.
- Nodes in the nodepool are preempted, rebooted, or transitioning to unhealthy state.
- Pods or container logs show collective communication timeouts (e.g., NCCL timeouts, MegaScale hangs).
- You need to identify if a specific physical host VM is consistently problematic.

## Components

- `SKILL.md`: Main instruction set for AI agents.
- `references/failure_signatures.md`: Example failure logs for pattern matching.
- `scripts/validate_queries.sh`: Syntax validator for diagnostic queries.
- `TEST.md`: Manual verification plan.
- `EVAL.textproto`: Evaluation suite for testing agent compliance.
- `BUILD`: Build definition for integration testing.

## Maintenance

LQL queries in `SKILL.md` should be periodically verified against Cloud Logging using `scripts/validate_queries.sh`, while MQL/PromQL queries should be validated against your Cloud Monitoring dashboard configurations to ensure they remain aligned with system schema updates.
