# Kubernetes Service Resource Schema

The `k8s_service` monitored resource represents a Kubernetes Service instance.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_service`
- Display name: Kubernetes Service
- Description: A Kubernetes Service instance.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location (region or zone).
- `cluster_name`: The cluster name.
- `namespace_name`: The namespace the service is running in.
- `service_name`: The service name.

## Sample queries

Service-level metrics often come from service mesh, ingress, or application
instrumentation. Replace the placeholder metric names below with the PromQL
metric names you collect for request count, latency, and error rate.

### Service request rate

```promql
topk(
  10,
  sum by (namespace_name, service_name) (
    rate(
      <request_count_metric>{
        monitored_resource="k8s_service",
        cluster_name="<cluster_name>"
      }[1m]
    )
  )
)
```

### Service request latency p95

```promql
topk(
  10,
  histogram_quantile(
    0.95,
    sum by (le, namespace_name, service_name) (
      rate(
        <request_latency_metric>_bucket{
          monitored_resource="k8s_service",
          cluster_name="<cluster_name>"
        }[5m]
      )
    )
  )
)
```

### Service error rate

```promql
topk(
  10,
  sum by (namespace_name, service_name) (
    rate(
      <error_count_metric>{
        monitored_resource="k8s_service",
        cluster_name="<cluster_name>"
      }[1m]
    )
  )
)
```

### Service throughput (bytes sent)

```promql
topk(
  10,
  sum by (namespace_name, service_name) (
    rate(
      <response_bytes_metric>{
        monitored_resource="k8s_service",
        cluster_name="<cluster_name>"
      }[1m]
    )
  )
)
```
