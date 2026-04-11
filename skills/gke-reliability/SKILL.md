---
name: gke-reliability
description: Workflows for ensuring high availability and reliability of GKE workloads.
---

# GKE Reliability Skill

This skill provides workflows for configuring your GKE cluster and workloads for high availability and reliability.

## Workflows

### 1. Verify Cluster High Availability

Check if the cluster is regional or has multi-zonal node pools.

**Command:**

```bash
gcloud container clusters describe <cluster-name> --region <region> --format="json(location, locations)"
```

If `location` is a region (e.g., `us-central1`), the control plane is regional.
If `locations` has multiple entries, nodes are spread across multiple zones.

### 2. Configure Pod Disruption Budgets (PDB)

PDBs ensure that a minimum number of pods are available during voluntary disruptions (like node upgrades).

**Check existing PDBs:**

```bash
kubectl get pdb -n <namespace>
```

**Example Manifest:**

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: my-app-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: my-app
```

### 3. Configure Health Probes

Ensure all production containers have Liveness, Readiness, and optionally Startup probes.

- **Readiness Probe**: Determines when a container is ready to start accepting traffic.
- **Liveness Probe**: Determines when to restart a container.
- **Startup Probe**: Disables liveness and readiness checks until the app has started up.

**Check workload probes:**

```bash
kubectl get deployment <app-name> -n <namespace> -o yaml | grep -E "livenessProbe|readinessProbe|startupProbe"
```

### 4. Graceful Shutdown

Ensure applications handle `SIGTERM` signals gracefully and have an appropriate `terminationGracePeriodSeconds` set (default is 30s).

### 5. Topology Spread Constraints

Ensure pods are spread across zones or nodes to avoid correlated failures.

**Example Manifest excerpt:**

```yaml
spec:
  topologySpreadConstraints:
    - maxSkew: 1
      topologyKey: topology.kubernetes.io/zone
      whenUnsatisfiable: DoNotSchedule # or ScheduleAnyway
      labelSelector:
        matchLabels:
          app: my-app
```

### 6. Maintenance Windows and Exclusions

Configure when GKE can perform automated upgrades to avoid peak hours.

**Command to set maintenance window:**

```bash
gcloud container clusters update <cluster-name> \
    --region <region> \
    --maintenance-window-start <start-time> \
    --maintenance-window-recurrence "FREQ=DAILY"
```

### 7. Multi-cluster Services (MCS)

For global availability and disaster recovery, use MCS to export and import services across clusters.

**Enable MCS in the fleet:**

```bash
gcloud container fleet multi-cluster-services enable \
    --project <project-id>
```

**Export a Service:**

```yaml
apiVersion: networking.gke.io/v1
kind: ServiceExport
metadata:
  name: my-service
  namespace: my-namespace
```

Applications in other clusters in the same fleet can then access this service using the `my-service.my-namespace.svc.clusterset.local` hostname.

## Best Practices

1. **Regional Clusters**: Always use regional clusters for production workloads to survive zone failures.
2. **Probes for All Containers**: Every container in a production pod should have at least a readiness probe.
3. **PDBs for Critical Apps**: Use PDBs to prevent downtime during automated node upgrades.
4. **Zone Spreading**: Always use `topologySpreadConstraints` to ensure pods are distributed across zones.
5. **Multi-cluster for DR**: Use Multi-cluster Services (MCS) and Gateway API for global traffic management and disaster recovery across regions.
6. **Schedule Maintenance**: Set maintenance windows to ensure upgrades happen during low-traffic periods.
