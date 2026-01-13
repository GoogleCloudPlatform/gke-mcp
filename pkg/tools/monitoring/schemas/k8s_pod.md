# Kubernetes Pod Resource Schema

The `k8s_pod` monitored resource represents a Kubernetes pod instance.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_pod`
- Display name: Kubernetes Pod
- Description: A Kubernetes pod instance.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.
- `namespace_name`: The namespace the pod is running in.
- `pod_name`: The pod name.

## Sample queries

### Top pods by network receive throughput

```promql
topk(
  10,
  sum by (namespace_name, pod_name) (
    rate(
      kubernetes_io:pod_network_received_bytes_count{
        monitored_resource="k8s_pod",
        cluster_name="<cluster_name>",
        interface="eth0"
      }[1m]
    )
  )
)
```

### Top pods by network send throughput

```promql
topk(
  10,
  sum by (namespace_name, pod_name) (
    rate(
      kubernetes_io:pod_network_sent_bytes_count{
        monitored_resource="k8s_pod",
        cluster_name="<cluster_name>",
        interface="eth0"
      }[1m]
    )
  )
)
```

### Pods with high volume utilization

```promql
topk(
  10,
  max by (namespace_name, pod_name) (
    kubernetes_io:pod_volume_utilization{
      monitored_resource="k8s_pod",
      cluster_name="<cluster_name>"
    }
  )
)
```

### Top pods by ephemeral storage usage

```promql
topk(
  10,
  avg by (namespace_name, pod_name) (
    kubernetes_io:pod_ephemeral_storage_used_bytes{
      monitored_resource="k8s_pod",
      cluster_name="<cluster_name>"
    }
  )
)
```

### Slowest pod startup (pod first ready latency p95)

```promql
topk(
  10,
  quantile_over_time(
    0.95,
    kubernetes_io:pod_latencies_pod_first_ready{
      monitored_resource="k8s_pod",
      cluster_name="<cluster_name>"
    }[5m]
  )
)
```

### Network policy events in the last 5 minutes

```promql
sum by (namespace_name, pod_name) (
  increase(
    kubernetes_io:pod_network_policy_event_count{
      monitored_resource="k8s_pod",
      cluster_name="<cluster_name>"
    }[5m]
  )
) > 0
```
