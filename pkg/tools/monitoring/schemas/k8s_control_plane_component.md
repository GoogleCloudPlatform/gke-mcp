# Kubernetes Control Plane Component Resource Schema

The `k8s_control_plane_component` monitored resource represents a GKE control plane component.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_control_plane_component`
- Display name: Kubernetes Control Plane Component
- Description: A Kubernetes Control Plane component.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location where the component runs.
- `cluster_name`: The cluster name.
- `component_name`: The control plane component name.
- `component_location`: The location where the component runs.

## Sample queries

Control plane metrics vary by configuration. Replace the placeholder metric
names below with the PromQL metric names available in your environment (for
example, API server request latency, request count, or error count metrics).

### API server request latency p95

```promql
histogram_quantile(
  0.95,
  sum by (le, cluster_name) (
    rate(
      <request_latency_metric>_bucket{
        monitored_resource="k8s_control_plane_component",
        component_name="kube-apiserver"
      }[5m]
    )
  )
)
```

### Control plane request rate by component

```promql
topk(
  5,
  sum by (component_name) (
    rate(
      <request_count_metric>{
        monitored_resource="k8s_control_plane_component"
      }[1m]
    )
  )
)
```

### Control plane error rate by component

```promql
topk(
  5,
  sum by (component_name) (
    rate(
      <error_count_metric>{
        monitored_resource="k8s_control_plane_component"
      }[1m]
    )
  )
)
```

### Component availability by location

```promql
min by (component_location, component_name) (
  <availability_metric>{monitored_resource="k8s_control_plane_component"}
)
```
