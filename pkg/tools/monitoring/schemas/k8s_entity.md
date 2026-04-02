# Kubernetes Entity Resource Schema

The `k8s_entity` monitored resource represents a Kubernetes entity such as a workload or namespace-scoped object.

See [Monitored resource types](https://cloud.google.com/monitoring/api/resources) for the full list of GKE resource schemas.

## Schema

- Resource type: `k8s_entity`
- Display name: Kubernetes Entity
- Description: A Kubernetes entity.

## Labels

- `project_id`: The GCP project ID associated with the cluster.
- `location`: The cluster location that contains the entity.
- `cluster_name`: The cluster name.
- `entity_type`: The entity type.
- `entity_namespace`: The entity namespace.
- `entity_name`: The entity name.
- `entity_uid`: The entity UID.

## Sample queries

### JobSet scheduling goodput (lowest)

```promql
bottomk(
  10,
  avg by (entity_namespace, entity_name) (
    kubernetes_io:jobset_scheduling_goodput{
      monitored_resource="k8s_entity",
      cluster_name="<cluster_name>"
    }
  )
)
```

### JobSet proxy runtime goodput (lowest)

```promql
bottomk(
  10,
  avg by (entity_namespace, entity_name) (
    kubernetes_io:jobset_proxy_runtime_goodput{
      monitored_resource="k8s_entity",
      cluster_name="<cluster_name>"
    }
  )
)
```

### Time between interruptions (p50)

```promql
histogram_quantile(
  0.50,
  sum by (le, entity_namespace, entity_name) (
    kubernetes_io:jobset_times_between_interruptions_bucket{
      monitored_resource="k8s_entity",
      cluster_name="<cluster_name>"
    }
  )
)
```

### Time to recover from interruptions (p95)

```promql
topk(
  10,
  histogram_quantile(
    0.95,
    sum by (le, entity_namespace, entity_name) (
      kubernetes_io:jobset_times_to_recover_bucket{
        monitored_resource="k8s_entity",
        cluster_name="<cluster_name>"
      }
    )
  )
)
```

### JobSet uptime by entity

```promql
topk(
  10,
  max by (entity_namespace, entity_name) (
    kubernetes_io:jobset_uptime{
      monitored_resource="k8s_entity",
      cluster_name="<cluster_name>"
    }
  )
)
```
