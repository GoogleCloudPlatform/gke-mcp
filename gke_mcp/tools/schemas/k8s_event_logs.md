# Kubernetes Event Logs Schema

In Kubernetes, Events are objects that provide information about resources,
such as state changes, node errors, Pod errors, or scheduling failures.
Various Kubernetes components, such as the kubelet or workload controllers,
create Events to report changes in objects. For example, the StatefulSet
controller might create an Event when the number of replicas in a StatefulSet
changes. For more information about Events, see the Event API
reference page and the Kubernetes glossary entry for Event.

See [GKE Event logging information](https://cloud.google.com/kubernetes-engine/docs/how-to/view-logs#k8s-event-logs)
for details about event logs on GKE.

## Schema

Note that k8s event logs are encoded into `LogEntry` objects.
The event information is encoded into a `jsonPayload` field.

The following are the most relevant fields in a Kubernetes event log entry:

- `insertId`: A unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. This value is always `projects/<project_id>/logs/events`
  where `<project_id>` is the ID of the project that owns the log entry.
- `receiveTimestamp`: The timestamp that the log entry was received by the
  logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the Monitored Resource.
  - `labels`:
    - `cluster_name`: The name of the Kubernetes cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
- `jsonPayload`: The payload of the log entry, containing the Kubernetes
  Event object in JSON format.
- `timestamp`: The timestamp of when the log entry was emitted.

## Sample Queries

### List event logs for one given cluster

This query lists all event logs for a given cluster, project, and location.

```lql
resource.type="k8s_cluster"
log_name="projects/<project_id>/logs/events"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```

### List event logs for one given cluster + kind

This query lists all event logs for a given cluster, project, location and kind.

```lql
resource.type="k8s_cluster"
log_name="projects/<project_id>/logs/events"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
jsonPayload.involvedObject.kind="<kind>"
```
