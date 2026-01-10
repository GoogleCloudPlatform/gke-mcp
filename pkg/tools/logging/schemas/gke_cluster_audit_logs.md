# GKE Cluster Audit Logs Schema

GKE emits Cloud Audit Logs for cluster operations such as creating clusters,
upgrading control planes, and modifying node pools. These logs help answer
who changed what, when, and whether the operation succeeded.

See [GKE audit logging](https://cloud.google.com/kubernetes-engine/docs/how-to/audit-logging)
for details.

## Schema

GKE cluster audit logs are encoded into `LogEntry` objects and use the
`gke_cluster` resource type.

The following are the most relevant fields in a GKE cluster audit log entry:

- `insertId`: A unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. Common values include:
  - `projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity` (Admin Activity)
  - `projects/<project_id>/logs/cloudaudit.googleapis.com%2Fsystem_event` (System Events)
- `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the monitored resource. For cluster audit logs, this is `gke_cluster`.
  - `labels`:
    - `cluster_name`: The name of the GKE cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
- `protoPayload`: The payload of the log entry, containing the audit information.
  - `@type`: The type of the proto payload. Always set to `type.googleapis.com/google.cloud.audit.AuditLog`.
  - `serviceName`: Typically `container.googleapis.com`.
  - `methodName`: The API method. For example, `google.container.v1.ClusterManager.UpdateCluster`.
  - `authenticationInfo.principalEmail`: The identity that initiated the request.
  - `status.code`: The status code of the request (non-zero indicates failure).
- `timestamp`: The timestamp of when the log entry was emitted.
- `severity`: The severity level of the log entry (DEBUG, INFO, WARNING, ERROR, etc.).

## Sample Queries

### Cluster operation logs for a cluster

```lql
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
resource.type="gke_cluster"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
```

### Cluster upgrade or control plane update operations

```lql
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
resource.type="gke_cluster"
resource.labels.cluster_name="<cluster_name>"
protoPayload.methodName="google.container.v1.ClusterManager.UpdateCluster"
```

### Node pool changes (upgrade, resize, or config update)

```lql
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
resource.type="gke_cluster"
resource.labels.cluster_name="<cluster_name>"
(protoPayload.methodName="google.container.v1.ClusterManager.UpdateNodePool"
 OR protoPayload.methodName="google.container.v1.ClusterManager.SetNodePoolSize"
 OR protoPayload.methodName="google.container.v1.ClusterManager.UpgradeNodePool")
```

### Failed cluster operations

```lql
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
resource.type="gke_cluster"
resource.labels.cluster_name="<cluster_name>"
protoPayload.status.code!=0
```

### System event logs for a cluster

```lql
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Fsystem_event"
resource.type="gke_cluster"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
```
