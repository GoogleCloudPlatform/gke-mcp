---
name: gke-multi-tenancy
description: Guidance on implementing multi-tenancy and governance in Google Kubernetes Engine (GKE) clusters.
---

# GKE Multi-tenancy and Governance

This skill provides guidance on implementing multi-tenancy and governance in Google Kubernetes Engine (GKE) clusters.

## Overview

Multi-tenancy allows you to share a single GKE cluster among multiple teams or applications securely. Governance ensures that policies and resource limits are enforced.

## Workflows

### 1. Create Namespaces for Isolation

Namespaces provide a scope for names and are the primary unit of isolation in Kubernetes.

**Steps:**

1. Create a namespace for each tenant.

**Example Namespace Manifest:**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: tenant-a
  labels:
    team: alpha
```

### 2. Configure RBAC for Least Privilege

Role-Based Access Control (RBAC) allows you to control who has access to what resources within a namespace.

**Steps:**

1. Define a `Role` with specific permissions.
2. Bind the `Role` to a user or group using a `RoleBinding`.

**Example Role and RoleBinding Manifest:**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: tenant-a
  name: pod-reader
rules:
  - apiGroups: [""] # "" indicates the core API group
    resources: ["pods"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: read-pods
  namespace: tenant-a
subjects:
  - kind: User
    name: user@example.com # Name is case sensitive
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: pod-reader
  apiGroup: rbac.authorization.k8s.io
```

### 3. Enforce Resource Quotas

Resource quotas prevent a single tenant from consuming all resources in the cluster.

**Example ResourceQuota Manifest:**

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: tenant-a-quota
  namespace: tenant-a
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
```

### 4. Hierarchical Namespaces (HNC)

HNC allows you to set up a hierarchy of namespaces, making it easier to manage policies and resource limits across multiple teams.

**Install HNC:**

```bash
kubectl apply -f https://github.com/kubernetes-sigs/hierarchical-namespaces/releases/latest/download/hnc-manager.yaml
```

**Create a Hierarchy:**

1. Create a parent namespace: `kubectl create ns parent-ns`
2. Create a child namespace: `kubectl create ns child-ns`
3. Set the parent:
   ```bash
   kubectl hnc set child-ns --parent parent-ns
   ```

**Propagate Resources:**
HNC automatically propagates certain resources (like Roles and RoleBindings) from parent to child namespaces.

## Best Practices

1. **Namespace Per Tenant**: Always use separate namespaces for different teams or applications.
2. **Hierarchical Namespaces (HNC)**: For complex organizations, use HNC to manage hierarchies of namespaces and propagate policies from parents to children.
3. **Least Privilege RBAC**: Grant only the permissions necessary for users and service accounts. Use `ClusterRole` and `RoleBinding` for consistency.
4. **Enforce Quotas**: Use Resource Quotas at both the namespace and (if using HNC) the parent level.
5. **Network Isolation**: Use Network Policies to enforce isolation between namespaces.
