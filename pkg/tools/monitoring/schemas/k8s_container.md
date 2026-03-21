# Kubernetes Container Resource Schema

The `k8s_container` monitored resource represents a container running in a Kubernetes pod.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_container`
- Display name: Kubernetes Container
- Description: A Kubernetes container instance.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.
- `namespace_name`: The namespace the pod is running in.
- `pod_name`: The pod name.
- `container_name`: The container name.

## Sample queries

### Top containers by CPU cores used

```promql
topk(
  10,
  avg by (namespace_name, pod_name, container_name) (
    rate(
      kubernetes_io:container_cpu_core_usage_time{
        monitored_resource="k8s_container",
        cluster_name="<cluster_name>"
      }[1m]
    )
  )
)
```

### Containers with high CPU request utilization

```promql
avg by (namespace_name, pod_name, container_name) (
  kubernetes_io:container_cpu_request_utilization{
    monitored_resource="k8s_container",
    cluster_name="<cluster_name>"
  }
) > 0.8
```

### Top containers by non-evictable memory usage

```promql
topk(
  10,
  avg by (namespace_name, pod_name, container_name) (
    kubernetes_io:container_memory_used_bytes{
      monitored_resource="k8s_container",
      cluster_name="<cluster_name>",
      memory_type="non-evictable"
    }
  )
)
```

### Containers with high memory limit utilization

```promql
avg by (namespace_name, pod_name, container_name) (
  kubernetes_io:container_memory_limit_utilization{
    monitored_resource="k8s_container",
    cluster_name="<cluster_name>"
  }
) > 0.8
```

### Containers restarting in the last 10 minutes

```promql
sum by (namespace_name, pod_name, container_name) (
  increase(
    kubernetes_io:container_restart_count{
      monitored_resource="k8s_container",
      cluster_name="<cluster_name>"
    }[10m]
  )
) > 0
```

### Top containers by ephemeral storage usage

```promql
topk(
  10,
  avg by (namespace_name, pod_name, container_name) (
    kubernetes_io:container_ephemeral_storage_used_bytes{
      monitored_resource="k8s_container",
      cluster_name="<cluster_name>"
    }
  )
)
```
