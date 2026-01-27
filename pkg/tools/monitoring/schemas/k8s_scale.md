# Kubernetes Scale Resource Schema

The `k8s_scale` monitored resource represents a Kubernetes object that can be targeted by autoscalers.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_scale`
- Display name: Kubernetes Scale
- Description: A Kubernetes object that can be targeted by Kubernetes autoscalers.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.
- `namespace_name`: The namespace containing the scaled object.
- `controller_api_group_name`: The API group of the scaled object, for example `core`.
- `controller_kind`: The kind of the scaled object, for example `Deployment`.
- `controller_name`: The name of the scaled object.

## Sample queries

### HPA recommended CPU request per replica

```promql
topk(
  10,
  avg by (namespace_name, controller_name) (
    kubernetes_io:autoscaler_container_cpu_per_replica_recommended_request_cores{
      monitored_resource="k8s_scale",
      cluster_name="<cluster_name>"
    }
  )
)
```

### HPA recommended memory request per replica

```promql
topk(
  10,
  avg by (namespace_name, controller_name) (
    kubernetes_io:autoscaler_container_memory_per_replica_recommended_request_bytes{
      monitored_resource="k8s_scale",
      cluster_name="<cluster_name>"
    }
  )
)
```

### HPA recommendation latency (p95)

```promql
topk(
  10,
  quantile_over_time(
    0.95,
    kubernetes_io:autoscaler_latencies_per_hpa_recommendation_scale_latency_seconds{
      monitored_resource="k8s_scale",
      cluster_name="<cluster_name>"
    }[5m]
  )
)
```
