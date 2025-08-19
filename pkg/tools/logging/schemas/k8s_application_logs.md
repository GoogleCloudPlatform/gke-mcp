# Kubernetes Application Logs Schema

Kubernetes containers collect logs for your workloads written to STDOUT and STDERR. You can find your workload application logs using the k8s_container resource type. Your logs will appear in Logging with the following schema.

## Schema

Note that k8s Application logs are encoded into `LogEntry` objects. The application information is encoded into a `jsonPayload` field.

The following are the most relevant fields in a Kubernetes Application log entry:

- `insertId`: An unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. This value is always `projects/<project_id>/logs/stderr` (logs written to standard error) or `projects/<project_id>logs/stdout` (logs written to standard out), where `<project_id>` is the ID of the project that owns the log entry.
- `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the Monitored Resource.
  - `labels`:
    - `cluster_name`: The name of the Kubernetes cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
    - `namespace_name`: The namespace of the GKE Workload
    - `pod_name`: The name of the GKE Pod
- `jsonPayload`: The text payload of the Application Logs
- `timestamp`: The timestamp of when the log entry was emitted.

## Sample Queries

### List k8s event logs written to standard error

This query lists all k8s event logs for a given cluster, project, and location.

```lql
resource.type="k8s_container"
logName="projects/<project_id>/logs/stderr"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```
