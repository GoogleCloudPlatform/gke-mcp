# ComputeClass CRD Specification

The `ComputeClass` resource (API group: `cloud.google.com/v1`) allows you to define custom node configurations and autoscaling priorities.

## Object Structure

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: <string> # Required. The name of the ComputeClass.
spec:
  # Required. Ordered list of rules. GKE attempts to satisfy them in order.
  priorities:
    - <PriorityRule>

  # Optional. Defaults to "DoNotScaleUp".
  # Options: "DoNotScaleUp", "ScaleUpAnyway"
  whenUnsatisfiable: <string>

  # Optional. Configuration for automatic node pool management.
  nodePoolAutoCreation:
    enabled: <boolean> # Default: true

  # Optional. Configuration for active migration of workloads.
  activeMigration:
    optimizeRulePriority: <boolean> # Default: false

  # Optional. Fine-tuning for autoscaling behavior.
  autoscalingPolicy:
    consolidationDelay: <duration> # e.g., "10m"

  # Optional. Default values for fields omitted in 'priorities'.
  priorityDefaults: <PriorityRule>
```

## PriorityRule Fields

Each item in `spec.priorities` (and `spec.priorityDefaults`) can contain the following fields:

| Field | Type | Description | Example |
|Ref|---|---|---|
| `machineFamily` | string | The Compute Engine machine family. | `n4`, `c3`, `t2a` |
| `machineType` | string | Specific machine type. | `n4-standard-32` |
| `spot` | boolean | Whether to use Spot VMs. | `true` |
| `minCores` | int | Minimum number of vCPUs. | `4` |
| `minMemoryGb` | int | Minimum memory in GB. | `16` |
| `gpu` | object | GPU configuration. | See below |
| `tpu` | object | TPU configuration. | See below |
| `storage` | object | Boot disk configuration. | See below |
| `location` | object | Zone/Region targeting. | See below |
| `reservations` | object | Consumption of specific reservations. | See below |

### GPU Configuration (`gpu`)

```yaml
gpu:
  type: <string> # e.g., "nvidia-l4", "nvidia-h100-80gb"
  count: <int> # Number of GPUs per node.
  driverVersion: <string> # Optional. e.g., "latest", "default"
```

### TPU Configuration (`tpu`)

```yaml
tpu:
  type: <string> # e.g., "v5p-slice"
  count: <int> # Number of TPU chips.
  topology: <string> # e.g., "2x2x1"
```

### Storage Configuration (`storage`)

```yaml
storage:
  bootDisk:
    type: <string> # e.g., "pd-balanced", "hyperdisk-balanced"
    sizeGb: <int> # Size of the boot disk.
    kmsKey: <string> # Optional. Cloud KMS key URI.
  localSsd:
    count: <int> # Number of local SSDs to attach.
    interface: <string> # Optional. "NVME" or "SCSI".
```

### Location Configuration (`location`)

```yaml
location:
  # List of specific zones to target.
  zones:
    - "us-central1-a"
    - "us-central1-b"

  # OR set type to "Any" to allow GKE to pick from cluster zones.
  type: "Any"
```

### Reservations (`reservations`)

```yaml
reservations:
  consumeReservationType: <string>
  # Options:
  # "NO_RESERVATION": Do not use reservations.
  # "ANY_RESERVATION": Use any applicable reservation.
  # "SPECIFIC_RESERVATION": Use a named reservation.

  key: <string> # "projects/.../reservations/name" (Required if SPECIFIC_RESERVATION)
  values: <list> # (Required if SPECIFIC_RESERVATION)
```

## Workload Usage

To use a `ComputeClass`, add a node selector to your Pod specification:

```yaml
nodeSelector:
  cloud.google.com/compute-class: "<compute-class-name>"
  cloud.google.com/gke-nodepool: "" # Optional but recommended to allow dynamic selection
```
