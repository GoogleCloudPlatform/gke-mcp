# Kubernetes Cluster Resource Schema

The `k8s_cluster` monitored resource represents a Kubernetes cluster in GKE.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_cluster`
- Display name: Kubernetes Cluster
- Description: A Kubernetes cluster. It contains Kubernetes audit logs from the cluster.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.

## Sample queries

Most GKE system metrics are emitted at node, pod, or container scope. The queries
below roll those metrics up to cluster-level views using the cluster labels.

### Cluster CPU allocatable utilization (node rollup)

```promql
topk(
  10,
  avg by (cluster_name) (
    kubernetes_io:node_cpu_allocatable_utilization{monitored_resource="k8s_node"}
  )
)
```

### Cluster memory usage (node rollup)

```promql
topk(
  10,
  sum by (cluster_name) (
    kubernetes_io:node_memory_used_bytes{
      monitored_resource="k8s_node",
      memory_type="non-evictable"
    }
  )
)
```

### Cluster pod network throughput (pod rollup)

```promql
topk(
  10,
  sum by (cluster_name) (
    rate(
      kubernetes_io:pod_network_received_bytes_count{
        monitored_resource="k8s_pod",
        interface="eth0"
      }[1m]
    )
  )
)
```

### Cluster container restarts in the last 10 minutes

```promql
sort_desc(
  sum by (cluster_name) (
    increase(
      kubernetes_io:container_restart_count{monitored_resource="k8s_container"}[10m]
    )
  )
)
```

### Cluster ephemeral storage usage (node rollup)

```promql
topk(
  10,
  sum by (cluster_name) (
    kubernetes_io:node_ephemeral_storage_used_bytes{monitored_resource="k8s_node"}
  )
)
```
