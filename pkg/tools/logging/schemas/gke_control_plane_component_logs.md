# GKE Control Plane Component Logs Schema

GKE control plane components (API server, scheduler, and controller manager)
emit logs that are critical for diagnosing API errors, scheduling failures,
and reconciliation issues.

See [GKE control plane component logs](https://cloud.google.com/kubernetes-engine/docs/how-to/view-logs#control_plane_logs)
for details.

## Schema

Control plane component logs are encoded into `LogEntry` objects and use the
`k8s_control_plane_component` resource type.

The following are the most relevant fields in a control plane component log entry:

- `insertId`: A unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. Common values include:
  - `projects/<project_id>/logs/container.googleapis.com%2Fapiserver`
  - `projects/<project_id>/logs/container.googleapis.com%2Fscheduler`
  - `projects/<project_id>/logs/container.googleapis.com%2Fcontroller-manager`
- `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the monitored resource. For control plane component logs,
    this is `k8s_control_plane_component`.
  - `labels`:
    - `cluster_name`: The name of the Kubernetes cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
    - `component_name`: The control plane component name (for example, `apiserver`).
- `jsonPayload` or `textPayload`: The payload of the log entry containing the log message.
- `timestamp`: The timestamp of when the log entry was emitted.
- `severity`: The severity level of the log entry (DEBUG, INFO, WARNING, ERROR, etc.).

## Sample Queries

### List all control plane component logs for a cluster

```lql
resource.type="k8s_control_plane_component"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```

### API server errors for a cluster

```lql
resource.type="k8s_control_plane_component"
logName="projects/<project_id>/logs/container.googleapis.com%2Fapiserver"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
severity>=ERROR
```

### Scheduler logs for failed scheduling and preemption

```lql
resource.type="k8s_control_plane_component"
logName="projects/<project_id>/logs/container.googleapis.com%2Fscheduler"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
(textPayload:"FailedScheduling" OR textPayload:"preemption" OR jsonPayload.message:"FailedScheduling")
```

### Controller manager errors for a cluster

```lql
resource.type="k8s_control_plane_component"
logName="projects/<project_id>/logs/container.googleapis.com%2Fcontroller-manager"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
severity>=ERROR
```
