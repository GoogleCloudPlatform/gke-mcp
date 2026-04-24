---
name: karpenter-to-ccc
description: Translates AWS EKS Karpenter NodePool and EC2NodeClass CRDs to GKE Custom Compute Class (CCC) configurations. You MUST use this skill whenever a user mentions migrating node provisioning from AWS/Karpenter to GKE, or provides Karpenter YAMLs and asks for GKE equivalents. This skill ensures complex mappings like Spot fallbacks, accelerator configurations, and system tunings are correctly handled.
---

# Karpenter to GKE ComputeClass Migration

## Instructions

### 1. Understand Core Differences

*   **Karpenter (EKS):** Provisions EC2 instances directly based on pending pod specifications and constraints defined in `NodePool` and `EC2NodeClass`. It dynamically selects instances from a wide range of possibilities to best fit the pod's needs at the time of scheduling.
*   **Custom Compute Classes (GKE):** Define a *profile* for node provisioning. This profile includes a prioritized list of machine configurations. GKE's Cluster Autoscaler, with Node Auto-Provisioning (NAP), uses the CCC to create and manage node pools, providing fallback mechanisms based on the defined priorities.

### 2. Analyze Input Karpenter CRDs

The user should provide the YAML definitions for their existing Karpenter `NodePool` and the associated `EC2NodeClass`.

**Key fields to analyze:**

*   **`NodePool.spec.template.spec.requirements`:** Node selection constraints (e.g., `node.kubernetes.io/instance-type`, `karpenter.sh/capacity-type`, `topology.kubernetes.io/zone`, custom labels).
*   **`NodePool.spec.template.spec.nodeClassRef`:** Reference to the `EC2NodeClass`.
*   **`NodePool.spec.template.metadata.labels`:** Labels to apply to NodeClaims and Nodes.
*   **`NodePool.spec.template.spec.taints` & `startupTaints`:** Taints to apply to nodes.
*   **`NodePool.spec.limits`:** Resource limits for the pool (e.g., `cpu`, `memory`).
*   **`NodePool.spec.disruption`:** Settings for `consolidationPolicy`, `consolidateAfter`, `expireAfter`.
*   **`NodePool.spec.weight`:** Priority of this NodePool.
*   **`EC2NodeClass.spec.amiFamily` & `amiSelectorTerms`:** How to select the node image.
*   **`EC2NodeClass.spec.role` / `instanceProfile`:** IAM role for the instances.
*   **`EC2NodeClass.spec.subnetSelectorTerms`:** Subnet selection for instances.
*   **`EC2NodeClass.spec.securityGroupSelectorTerms`:** Security group selection.
*   **`EC2NodeClass.spec.blockDeviceMappings`:** EBS volume configuration.
*   **`EC2NodeClass.spec.userData`:** Custom startup script.
*   **`EC2NodeClass.spec.metadataOptions`:** IMDS configuration.
*   **`EC2NodeClass.spec.kubelet`:** Kubelet arguments.
*   **`EC2NodeClass.spec.tags`:** Tags for AWS resources.

### 3. Map Karpenter Fields to GKE ComputeClass Fields

Translate the settings into the GKE `ComputeClass` CRD structure.

| Karpenter `NodePool` / `EC2NodeClass` Field | GKE ComputeClass Equivalent (`spec`) | Notes |
| :--- | :--- | :--- |
| `NodePool.spec.template.spec.requirements` (Instance Type, Size, Family, Arch) | `priorities[]` array: `machineFamily`, `minCores`, `minMemoryGb`, `arch` (e.g., `X86_64`, `ARM64`). | Define multiple `priorities` entries to simulate fallbacks. Assign `priorityScore` values (GKE 1.35+) or rely on array order. |
| `NodePool.spec.template.spec.requirements` (`karpenter.sh/capacity-type`: `spot` / `on-demand`) | `priorities[].spot`: `true` or `false` | List Spot options with higher priority for cost optimization. |
| `NodePool.spec.template.spec.requirements` (`topology.kubernetes.io/zone`) | `priorities[].location.zones` | Specify allowed GCP zones. GKE 1.33+ |
| `NodePool.spec.template.metadata.labels` | `nodePoolConfig.nodeLabels` (global) or `priorities[].nodeLabels` (specific) | Labels applied to nodes. Priority-level labels require GKE 1.34+ |
| `NodePool.spec.template.spec.taints` | `nodePoolConfig.taints` (global) or `priorities[].taints` (specific) | Taints applied to nodes. Priority-level taints require GKE 1.34+ |
| `NodePool.spec.template.spec.startupTaints` | No direct equivalent. Use `taints`. | GKE's model doesn't explicitly distinguish startup taints in CCC. |
| `EC2NodeClass.spec.role` / `instanceProfile` | `nodePoolConfig.serviceAccount` | Map AWS IAM Role to GCP Service Account. |
| `EC2NodeClass.spec.amiFamily` / `amiSelectorTerms` | `nodePoolConfig.imageType` (e.g., `COS_CONTAINERD`, `UBUNTU_CONTAINERD`) | GKE uses managed images. Use `nodePoolConfig.imageStreaming` for faster startups. |
| `EC2NodeClass.spec.blockDeviceMappings` | `priorities[].storage` (`bootDiskSize`, `bootDiskType`, `localSSDCount`) | GKE supports boot disk customization and Local SSDs in CCC. Secondary disks map to `secondaryBootDisks`. |
| `EC2NodeClass.spec.userData` | `nodePoolConfig.nodeSystemConfig.linuxNodeConfig.sysctls` | Map kernel tunings to `sysctls`. Other logic should move to DaemonSets. |
| `EC2NodeClass.spec.securityGroupSelectorTerms` | VPC Firewall Rules & Network Policies. | Network-level mapping. |
| `EC2NodeClass.spec.subnetSelectorTerms` | GKE Cluster's Node Subnet. | Use `location.zones` to influence subnet placement. |
| `EC2NodeClass.spec.tags` | `nodePoolConfig.resourceManagerTags` (GCP tags). | Map AWS tags to GCP Resource Manager Tags. |
| `NodePool.spec.disruption.consolidationPolicy` | `autoscalingPolicy.consolidationThreshold` | Tunes GKE's node consolidation behavior. |
| `NodePool.spec.disruption.expireAfter` | `priorities[].maxRunDurationSeconds` | Map node TTL to max run duration. GKE 1.32+ |
| `EC2NodeClass.spec.kubelet` | `nodePoolConfig.nodeSystemConfig.kubeletConfig` | Map Kubelet args like `cpuCfsQuota`, `podPidsLimit`. |
| `NodePool.spec.weight` | `priorities[].priorityScore` | Explicitly rank priorities. GKE 1.35+ |
| `Requirements` (GPU types) | `priorities[].gpu` (type, count) or `priorities[].tpu` | Direct accelerator mapping including GPU sharing (MPS/Time-sharing). |
| `Topology` (Compact Placement) | `priorities[].placement.policyName` | Map cluster-placement constraints to GCP Resource Policies. GKE 1.33+ |

### 4. Machine Family Translation:

*   AWS `m6i/m7i` -> GCP `n2`, `n4`
*   AWS `c6i/c7i` -> GCP `c3`, `c4`
*   AWS `r6i/r7i` -> GCP `n2-highmem`
*   AWS `t3` -> GCP `e2`
*   AWS `arm64` (Graviton) -> GCP `t2a`
*   AWS `p4d/p5` -> GCP `a2` (L4), `a3` (H100)

### 5. Construct the GKE ComputeClass YAML

```yaml
apiVersion: cloud.google.com/v1
kind: ComputeClass
metadata:
  name: # e.g., my-migrated-class
spec:
  nodePoolAutoCreation:
    enabled: true
  whenUnsatisfiable: ScaleUpAnyway # Or DoNotScaleUp

  activeMigration:
    optimizeRulePriority: true # Reconcile to higher priority nodes
    ensureAllDaemonSetPodsRunning: true # Scale up if DaemonSets don't fit

  autoscalingPolicy:
    consolidationDelayMinutes: 15
    consolidationThreshold: 70
    # gpuConsolidationThreshold: 100

  nodePoolConfig:
    imageType: COS_CONTAINERD
    serviceAccount: <SA>@<PROJECT>.iam.gserviceaccount.com
    nodeLabels:
      migrated-from: karpenter
    # imageStreaming:
    #   enabled: true

  priorities:
    # Priority 1: Spot N4 with local SSDs
    - priorityScore: 100
      machineFamily: n4
      spot: true
      storage:
        bootDiskSize: 100
        bootDiskType: hyperdisk-balanced
        localSSDCount: 1
      nodeLabels:
        capacity-type: spot
      nodeSystemConfig:
        linuxNodeConfig:
          sysctls:
            net.core.somaxconn: 2048

    # Priority 2: On-demand N4 (Fallback)
    - priorityScore: 90
      machineFamily: n4
      spot: false
      storage:
        bootDiskSize: 100

    # Priority 3: GPU Fallback (if applicable)
    # - priorityScore: 80
    #   gpu:
    #     type: nvidia-l4
    #     count: 1
    #     driverVersion: default

    # Priority 4: Specified Location
    # - priorityScore: 70
    #   machineFamily: e2
    #   location:
    #     zones: ["us-central1-a"]
```

### 6. Validation & Best Practices
*   **Active Migration:** Enable `optimizeRulePriority: true` to ensure workloads eventually land on your most preferred (e.g., Spot) nodes.
*   **Consolidation:** Match `consolidationThreshold` to your Karpenter `consolidationPolicy`.
*   **Accelerators:** For GPU workloads, specify `gpu.type` and `gpu.count` explicitly in a priority block.
*   **Network:** Use `gvnic: { enabled: true }` in `nodePoolConfig` for high-performance networking on supported machine types.
*   **Service Accounts:** Ensure the GCP Service Account has `roles/container.nodeServiceAccount`.
