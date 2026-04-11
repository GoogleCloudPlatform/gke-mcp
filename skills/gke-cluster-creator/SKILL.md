---
name: gke-creation
description: Guides the user through creating GKE clusters using pre-defined templates (Standard, Autopilot, GPU/AI).
---

# GKE Cluster Creation Skill

This skill helps users create Google Kubernetes Engine (GKE) clusters by providing a set of best-practice templates and guiding them through the customization process.

## core_behavior

1. **Template Selection**:
   - Present the available templates to the user if they haven't specified one.
   - Explain the trade-offs (e.g., Cost vs. Availability, Autopilot vs. Standard).
2. **Customization**:
   - Once a template is selected, present the default configuration (JSON/YAML).
   - Ask the user for essential missing information: `project_id`, `location`, `cluster_name`.
   - Ask if they want to modify optional fields (e.g., `machineType`, `nodeCount`, `network`).
3. **Validation**:
   - Ensure `project_id`, `location`, and `cluster_name` are set.
   - Ensure the configuration matches the `create_cluster` MCP tool schema.
4. **Execution**:
   - Call the `create_cluster` MCP tool with the final configuration.

## best_practices

When guiding the user or generating configurations, adhere to the following GKE cluster creation best practices:

### Security

1. **Private Clusters**: Default to private clusters with a private control plane and restricted public endpoints to minimize attack surface.
2. **VPC-Native Networking**: Use VPC-native clusters to enable alias IP ranges, which allows pod-level firewall rules and better network security.
3. **Workload Identity**: Prefer Workload Identity for securely granting GKE workloads access to Google Cloud services instead of using static service account keys.
4. **Shielded GKE Nodes**: Enable Shielded GKE Nodes to protect against rootkits and bootkits.
5. **Least Privilege (RBAC)**: Institute strict Role-Based Access Control limits granting minimal privilege to users and workloads.

### Cost Optimization

1. **Autoscaling**: Enable Cluster Autoscaler and Horizontal Pod Autoscaler to adjust resources based on demand.
2. **Right-Sizing**: Choose the appropriate machine types and node counts. Consider Spot VMs for fault-tolerant, non-critical workloads.

### High Availability & Reliability

1. **Regional Clusters**: Use Regional Clusters for production environments to ensure control plane replication across multiple zones. (Note: standard regional creates nodes across 3 zones by default).
2. **Pod Disruption Budgets**: Recommend setting Pod Disruption Budgets for application stability during node maintenance.
3. **Release Channels**: Subscribe to a release channel (e.g., Regular or Stable) for automated and safer cluster upgrades.

## templates

### 1. Autopilot (Recommended / Operations-Free)

Best for: Most workloads where you don't want to manage nodes. Secure and cost-optimized by default.

```json
{
  "name": "projects/{PROJECT_ID}/locations/{REGION}/clusters/{CLUSTER_NAME}",
  "autopilot": {
    "enabled": true
  }
}
```

### 2. Standard Regional (High Availability)

Best for: Production workloads requiring specific node configurations or control over node versioning.

```json
{
  "name": "projects/{PROJECT_ID}/locations/{REGION}/clusters/{CLUSTER_NAME}",
  "initialNodeCount": 1,
  "nodeConfig": {
    "machineType": "e2-standard-4",
    "diskSizeGb": 100,
    "oauthScopes": ["https://www.googleapis.com/auth/cloud-platform"]
  }
}
```

### 3. AI Inference (L4 with Model Caching)

Best for: High-performance AI/ML Inference. Uses secondary boot disks for faster model loading.

```json
{
  "name": "projects/{PROJECT_ID}/locations/{REGION}/clusters/{CLUSTER_NAME}",
  "initialNodeCount": 1,
  "nodeConfig": {
    "machineType": "g2-standard-4",
    "accelerators": [
      {
        "acceleratorCount": "1",
        "acceleratorType": "nvidia-l4"
      }
    ],
    "bootDiskKmsKey": "{KMS_KEY}",
    "additionalBootDisks": [
      {
        "sourceImage": "projects/deeplearning-platform-release/global/images/family/common-cu121-v20240111",
        "diskSizeGb": 100
      }
    ],
    "oauthScopes": ["https://www.googleapis.com/auth/cloud-platform"]
  }
}
```

### 4. Standard Zonal (Cost-Effective Dev/Test)

Best for: Development, testing, non-critical workloads.

```json
{
  "name": "projects/{PROJECT_ID}/locations/{ZONE}/clusters/{CLUSTER_NAME}",
  "initialNodeCount": 1,
  "nodeConfig": {
    "machineType": "e2-medium",
    "diskSizeGb": 50,
    "oauthScopes": ["https://www.googleapis.com/auth/cloud-platform"]
  }
}
```

## instructions

- **PRIORITIZE** Autopilot for almost all user requests. Only suggest Standard if they need specific node-level features (e.g., specific kernel versions, local SSD customizations beyond what Autopilot offers).
- **ALWAYS** ask for the `project_id` if it is not in the context.
- **ALWAYS** ask for the `location` (Region or Zone).
- **ALWAYS** ask for a unique `cluster_name`.
- **USE** `create_cluster` MCP tool to create the cluster.

## example_usage

**User**: "I want to create a GKE cluster."
**Model**: "I can help with that. What kind of cluster do you need?

1. **Standard Zonal**: Good for dev/test.
2. **Standard Regional**: High availability.
3. **Autopilot**: Fully managed.
4. **GPU Enabled**: For AI/ML workloads."

**User**: "Standard Zonal, please."
**Model**: "Great. I'll need a few details:

- Project ID
- Zone (e.g., us-central1-a)
- Cluster Name"

**User**: "Project `my-proj`, zone `us-west1-b`, name `dev-cluster`."
**Model**: "Here is the configuration I will use:
[JSON view]
Do you want to proceed?"
