# GKE Cluster Lifecycle and Upgrades

This skill provides guidance on managing the lifecycle and upgrades of Google Kubernetes Engine (GKE) clusters.

## Overview

Managing cluster upgrades is crucial for security and access to new features. GKE provides automated upgrades, but they must be configured to minimize disruption.

## Workflows

### 1. Select Release Channels

Release channels allow you to choose the balance between stability and feature availability.

- **Rapid**: Newest features, less tested.
- **Regular** (Default): Good balance.
- **Stable**: Most tested, best for critical production workloads.

**Command to set release channel:**

```bash
gcloud container clusters update <cluster-name> \
    --release-channel=stable \
    --region <region>
```

### 2. Configure Surge Upgrades

Surge upgrades allow you to specify how many nodes can be created above the target size during an upgrade, minimizing disruption.

**Example configuration:**

```bash
gcloud container node-pools update <pool-name> \
    --cluster=<cluster-name> \
    --max-surge-upgrade=2 \
    --max-unavailable-upgrade=0 \
    --region <region>
```

Setting `max-unavailable-upgrade=0` ensures that no nodes are taken offline before new ones are ready.

### 3. Implement Blue/Green Node Pool Upgrades

For high-risk upgrades, you can create a new node pool (Green) with the new version, test it, and then migrate workloads from the old node pool (Blue).

**Steps:**

1. Create a new node pool with the new version and appropriate taints (using --node-taints).
2. Cordon and drain the old node pool gradually.
3. Delete the old node pool once empty.

### 4. GitOps with Config Sync

Use Config Sync to manage your cluster configuration (e.g., RBAC, Namespaces, NetworkPolicies) from a Git repository.

**Enable Config Sync:**

```bash
gcloud container fleet config-management enable
# Configure for your cluster
gcloud container fleet config-management apply \
    --membership=<membership-name> \
    --config-sync-git-repo=<git-url> \
    --config-sync-git-branch=main \
    --config-sync-git-dir=/ \
    --config-sync-git-auth=none
```

**Benefits:**
- **Source of Truth**: Cluster state is defined in Git.
- **Drift Correction**: Config Sync automatically corrects manual changes that deviate from the Git repo.
- **Multi-cluster Consistency**: Easily apply the same configuration across multiple clusters.

## Best Practices

1. **Use Autopilot**: For most workloads, use **GKE Autopilot**, which automates node management, upgrades, and security hardening.
2. **Use Release Channels**: Always enroll production clusters in a release channel (preferably `Stable` or `Regular`).
3. **Configure Surge Upgrades**: Use `max-surge-upgrade` to ensure availability during upgrades.
4. **Use Maintenance Windows & Exclusions**: Configure maintenance windows to ensure upgrades only happen during off-peak hours, and use maintenance exclusions to prevent upgrades during critical business events (see [gke-reliability](../gke-reliability/SKILL.md)).
5. **Test in Non-Prod**: Always test upgrades in a staging environment before applying them to production.
