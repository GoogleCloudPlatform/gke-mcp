# Kubernetes Audit Logs Schema

The Kubernetes API server emits logs for each request that it processes, according to GKE’s audit policy.
These logs are useful to understand which user performed what operation through the Kubernetes API server.

See [GKE audit logging information](https://cloud.google.com/kubernetes-engine/docs/how-to/audit-logging) for details about Kubernetes audit logs on GKE.

## Schema

Kubernetes audit logs are encoded into `LogEntry` objects. The audit information is encoded into a `protoPayload` field.

The following are the most relevant fields in a Kubernetes audit log entry:

- `insertId`: A unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. Common values include:
  - `projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity` (Admin Activity, typically write requests)
  - `projects/<project_id>/logs/cloudaudit.googleapis.com%2Fdata_access` (Data Access, typically read requests, if enabled)
- `timestamp`: The timestamp of when the log entry was emitted.
- `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the Monitored Resource. For Kubernetes audit logs, this is always `k8s_cluster`.
  - `labels`:
    - `cluster_name`: The name of the Kubernetes cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
- `protoPayload`: The payload of the log entry, containing the audit information.
  - `@type`: The type of the proto payload. Always set to `type.googleapis.com/google.cloud.audit.AuditLog`.
  - `serviceName`: This value is always `k8s.io`.
  - `methodName`: The name of the Kubernetes API method. Formatted as `io.k8s.<api_group>.<api_version>.<resource>.<verb>`. For example, `io.k8s.core.v1.configmaps.get`.
  - `resourceName`: The name of the Kubernetes resource. Formatted as `<api_group>/<api_version>/namespaces/<namespace>/<resource-name>` for namespaced resources, or `<api_group>/<api_version>/<resource-name>` for cluster-scoped resources. For example, `core/v1/namespaces/foo/configmaps/my-configmap`.
  - `authenticationInfo.principalEmail`: The identity that initiated the request.
  - `authorizationInfo`: Authorization checks, including permissions and granted/denied decisions.
  - `requestMetadata.callerIp`: The client IP that initiated the request.
  - `responseStatus.code`: The HTTP status code for the request (if present).

## Sample Queries

### List data access audit logs (if enabled)

This query lists all data access audit logs for a given cluster, project, and location.

```lql
resource.type="k8s_cluster"
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Fdata_access"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```

### List write operations (create/update/delete)

```lql
resource.type="k8s_cluster"
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
resource.labels.cluster_name="<cluster_name>"
(protoPayload.methodName:".create"
 OR protoPayload.methodName:".update"
 OR protoPayload.methodName:".delete")
```

### Track RBAC changes by non-system users

```lql
resource.type="k8s_cluster"
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
protoPayload.methodName:"io.k8s.authorization.rbac"
NOT protoPayload.authenticationInfo.principalEmail:"system"
```

### Audit changes to a specific workload

```lql
resource.type="k8s_cluster"
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
protoPayload.resourceName:"namespaces/<namespace>/pods/<pod_name>"
```

### Find non-system access to the API server

```lql
resource.type="k8s_cluster"
logName="projects/<project_id>/logs/cloudaudit.googleapis.com%2Factivity"
NOT protoPayload.authenticationInfo.principalEmail:"system"
```
