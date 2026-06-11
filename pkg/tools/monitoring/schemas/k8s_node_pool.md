# Kubernetes Node Pool Resource Schema

The `k8s_node_pool` monitored resource represents a GKE node pool.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_node_pool`
- Display name: Kubernetes Nodepool
- Description: A Kubernetes nodepool instance.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.
- `node_pool_name`: The node pool name.

## Sample queries

### Node pool interruption count (last hour)

```promql
sort_desc(
  sum by (node_pool_name) (
    sum_over_time(
      kubernetes_io:node_pool_interruption_count{
        monitored_resource="k8s_node_pool",
        cluster_name="<cluster_name>"
      }[1h]
    )
  )
)
```

### Node pool status overview

```promql
max by (node_pool_name) (
  kubernetes_io:node_pool_status{
    monitored_resource="k8s_node_pool",
    cluster_name="<cluster_name>"
  }
)
```

### Accelerator time to recover (p95)

```promql
topk(
  10,
  histogram_quantile(
    0.95,
    sum by (le, node_pool_name) (
      kubernetes_io:node_pool_accelerator_times_to_recover_bucket{
        monitored_resource="k8s_node_pool",
        cluster_name="<cluster_name>"
      }
    )
  )
)
```

### Multi-host availability by node pool

```promql
min by (node_pool_name) (
  kubernetes_io:node_pool_multi_host_available{
    monitored_resource="k8s_node_pool",
    cluster_name="<cluster_name>"
  }
)
```
