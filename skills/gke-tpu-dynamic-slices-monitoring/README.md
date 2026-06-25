# Skill: GKE TPU Dynamic Slices Monitoring & Management

This skill provides a structured workflow for monitoring GKE TPU Dynamic Slices custom resources, deploying single and multi-slice workloads, diagnosing provisioning and creation failures, and decommissioning the Slice Controller.

## What is GKE TPU Dynamic Slices?

Dynamic Slices allow GKE users to request subsets or partitions of TPU topologies dynamically using a custom scheduler and `Slice` custom resources. GKE provisions and manages the lifecycle of these slice allocations.

## When to use this skill

Use this skill when:

- Monitoring the status of `Slice` resources (`accelerator.gke.io/v1beta1`) in a GKE cluster.
- Investigating failed slice creations (e.g. status shows `SliceCreationFailed`, `FAILED`, `ACTIVE_DEGRADED`).
- Authoring or validating Kubernetes Job or JobSet manifests for dynamic TPU slice deployment.
- Safely deleting dynamic slices or disabling the Slice Controller.

## Components

- `SKILL.md`: Main instructions and diagnostic workflow.
- `references/failure_signatures.md`: Real-world and expected status condition examples for pattern calibration.
- `scripts/validate_queries.sh`: Syntax validator for diagnostic queries.
- `TEST.md`: Manual verification plan.
- `EVAL.yaml`: Evaluation suite for performance tracking.

## Maintenance

Keep the diagnostic queries and commands aligned with GKE `accelerator.gke.io` API versions.
