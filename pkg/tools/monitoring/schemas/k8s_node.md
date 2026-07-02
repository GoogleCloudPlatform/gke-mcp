# Kubernetes Node Resource Schema

The `k8s_node` monitored resource represents a Kubernetes node instance.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_node`
- Display name: Kubernetes Node
- Description: A Kubernetes node instance.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.
- `node_name`: The node name.

## Sample queries

### Top nodes by CPU cores used

```promql
topk(
  10,
  avg by (node_name) (
    rate(
      kubernetes_io:node_cpu_core_usage_time{
        monitored_resource="k8s_node",
        cluster_name="<cluster_name>"
      }[1m]
    )
  )
)
```

### Nodes with high CPU allocatable utilization

```promql
topk(
  10,
  avg by (node_name) (
    kubernetes_io:node_cpu_allocatable_utilization{
      monitored_resource="k8s_node",
      cluster_name="<cluster_name>"
    }
  )
)
```

### Top nodes by non-evictable memory usage

```promql
topk(
  10,
  avg by (node_name) (
    kubernetes_io:node_memory_used_bytes{
      monitored_resource="k8s_node",
      cluster_name="<cluster_name>",
      memory_type="non-evictable"
    }
  )
)
```

### Top nodes by network receive throughput

```promql
topk(
  10,
  sum by (node_name) (
    rate(
      kubernetes_io:node_network_received_bytes_count{
        monitored_resource="k8s_node",
        cluster_name="<cluster_name>"
      }[1m]
    )
  )
)
```

### Top nodes by ephemeral storage usage

```promql
topk(
  10,
  avg by (node_name) (
    kubernetes_io:node_ephemeral_storage_used_bytes{
      monitored_resource="k8s_node",
      cluster_name="<cluster_name>"
    }
  )
)
```

### Nodes interrupted in the last hour

```promql
sum by (node_name) (
  max_over_time(
    kubernetes_io:node_interruption_count{
      monitored_resource="k8s_node",
      cluster_name="<cluster_name>"
    }[1h]
  )
) > 0
```
