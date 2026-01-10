# GKE Control Plane Access Logs Schema

If you enable GKE control plane access logs, GKE records incoming network
connections to control plane instances and SSH events for the control plane.
These logs are useful for verifying administrative access and correlating
with Access Transparency.

See [GKE control plane access logs](https://cloud.google.com/kubernetes-engine/docs/how-to/view-logs#control-plane-access-logs)
for details.

## Schema

Control plane access logs are encoded into `LogEntry` objects and use the
`gke_cluster` resource type.

The following are the most relevant fields in a control plane access log entry:

- `insertId`: A unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. Common values include:
  - `projects/<project_id>/logs/container.googleapis.com%2Fkcp_connection`
  - `projects/<project_id>/logs/container.googleapis.com%2Fkcp_ssh`
- `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the monitored resource. For control plane access logs,
    this is `gke_cluster`.
  - `labels`:
    - `cluster_name`: The name of the GKE cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
- `jsonPayload` or `textPayload`: The payload of the log entry containing access details.
- `timestamp`: The timestamp of when the log entry was emitted.
- `severity`: The severity level of the log entry (DEBUG, INFO, WARNING, ERROR, etc.).

## Sample Queries

### List control plane connection logs for a cluster

```lql
resource.type="gke_cluster"
logName="projects/<project_id>/logs/container.googleapis.com%2Fkcp_connection"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
```

### List control plane SSH access logs for a cluster

```lql
resource.type="gke_cluster"
logName="projects/<project_id>/logs/container.googleapis.com%2Fkcp_ssh"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
```

### Flag denied or failed control plane connections

```lql
resource.type="gke_cluster"
logName="projects/<project_id>/logs/container.googleapis.com%2Fkcp_connection"
resource.labels.cluster_name="<cluster_name>"
(textPayload:"denied" OR textPayload:"failed" OR jsonPayload.message:"denied" OR jsonPayload.message:"failed")
```
