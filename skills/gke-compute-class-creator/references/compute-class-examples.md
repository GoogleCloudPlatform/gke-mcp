# ComputeClass Examples

This file contains common patterns for defining GKE ComputeClasses.

## Scenario 1: Spot VMs with Fallback to On-Demand

This configuration prioritizes Spot VMs for cost savings but falls back to standard On-Demand VMs if Spot capacity is unavailable.

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: spot-with-fallback
spec:
  nodePoolAutoCreation:
    enabled: true
  priorities:
    # Priority 1: Try to get Spot VMs in the N4 family
    - machineFamily: n4
      spot: true
    # Priority 2: Fallback to On-Demand VMs in the N4 family
    - machineFamily: n4
      spot: false
```

## Scenario 2: GPU Workload (L4)

This configuration targets NVIDIA L4 GPUs for AI inference workloads.

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: l4-gpu-class
spec:
  priorities:
    - machineFamily: g2
      gpu:
        type: nvidia-l4
        count: 1
      minCores: 4
      minMemoryGb: 16
      storage:
        bootDisk:
          type: pd-balanced
          sizeGb: 100
```

## Scenario 3: High Performance Compute (C3)

Prioritizes the C3 machine family for compute-intensive workloads.

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: high-perf-compute
spec:
  priorities:
    - machineFamily: c3
      minCores: 8
      minMemoryGb: 16
```

## Scenario 4: Specific Zone Targeting

Ensures workloads only run in specific zones (e.g., to colocatee with other zonal resources).

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: zonal-targeting
spec:
  priorities:
    - machineFamily: e2
      location:
        zones:
          - us-central1-a
          - us-central1-b
```

## Scenario 5: Active Migration

Enables active migration to move workloads back to higher-priority resources (e.g., Spot VMs) when they become available.

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: active-migration-enabled
spec:
  activeMigration:
    optimizeRulePriority: true
  priorities:
    - machineFamily: n2
      spot: true
    - machineFamily: n2
      spot: false
```
