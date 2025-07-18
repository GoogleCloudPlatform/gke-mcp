# Kubernetes Audit Logs Schema

The Kubernetes API server emits logs for each request that it processes, according to GKE’s audit policy. 
These logs are useful to understand which user performed what operation through the Kubernetes API server.

See https://cloud.google.com/kubernetes-engine/docs/how-to/audit-logging for details about Kubernetes audit logs on GKE.

## Schema

Note that k8s audit logs are encoded into `LogEntry` objects. The audit information is encoded into a `protoPayload` field.

The following are the most relevant fields in a Kubernetes audit log entry:

-   `insertId`: An unique, auto-generated ID for the log entry.
-   `logName`: The name of the log entry. This value is always `projects/<project_id>/logs/cloudaudit.googleapis.com%2Fdata_access` (if the request is read-only) or `projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity` (if the request is write-only), where `<project_id>` is the ID of the project that owns the log entry.
-   `operation`: Information about an operation associated with the log entry, if applicable.
    -   `id`: The ID of the long-running operation that this log entry is associated with. This field has the value of `insertId`, because Kubernetes does not support long-running operations.
    -   `first`: Always set to `true`. Indicates whether this audit log entry is for the request that originated a long-running operation.
    -   `last`: Always set to `true`. Indicates whether this audit log entry is for the request that completed a long-running operation.
    -   `producer`: Always set to `k8s.io`.
-   `timestamp`: The timestamp of when the log entry was emitted.
-   `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
-   `resource`: The monitored resource that the log entry is associated with.
    -   `type`: The type of the Monitored Resource. For Kubernetes audit logs, this is always `k8s_cluster`.
    -   `labels`:
        -   `cluster_name`: The name of the Kubernetes cluster.
        -   `project_id`: The ID of the GCP project where the GKE cluster is located.
        -   `location`: The location of the GKE cluster (region or zone).
-   `protoPayload`: The payload of the log entry, containing the audit information.
    -   `@type`: The type of the proto payload. Always set to `type.googleapis.com/google.cloud.audit.AuditLog`.
    -   `serviceName`: This value is always `k8s.io`.
    -   `methodName`: The name of the Kubernetes API method. Formatted as `io.k8s.<api_group>.<api_version>.<resource>.<verb>`. For example, `io.k8s.core.v1.configmaps.get`.
    -   `resourceName`: The name of the Kubernetes resource. Formatted as `<api_group>/<api_version>/namespaces/<namespace>/<resource-name>` for namespaced resources, or `<api_group>/<api_version>/<resource-name>` for cluster-scoped resources. For example, `core/v1/namespaces/foo/configmaps/my-configmap`.

## Sample Queries

### List data access audit logs

This query lists all data access audit logs for a given cluster, project, and location.

```lql
resource.type="k8s_cluster"
log_name="projects/<project_id>/logs/cloudaudit.googleapis.com%2Fdata_access"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```
