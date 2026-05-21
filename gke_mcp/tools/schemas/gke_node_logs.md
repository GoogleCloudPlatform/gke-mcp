# GKE Node System Logs Schema

GKE nodes run several system-level components that are not containerized. These include
the container runtime, kubelet, and various monitoring daemons. Logs from these components
are essential for debugging node-level issues such as kubelet failures, container runtime
errors, and node health problems.

See [GKE system logging](https://cloud.google.com/kubernetes-engine/docs/how-to/view-logs)
for details about node system logs on GKE.

## System Components

The following non-containerized system components emit logs on GKE nodes:

- `kubelet`: The primary node agent that manages pods and containers
- `containerd` / `docker`: The container runtime
- `kubelet-monitor`: Monitors kubelet health
- `node-problem-detector`: Detects node problems like hardware failures and kernel issues
- `kube-container-runtime-monitor`: Monitors the container runtime health

## Schema

Note that node system logs are encoded into `LogEntry` objects. The log content is
typically encoded into a `jsonPayload` or `textPayload` field depending on the component.

The following are the most relevant fields in a GKE node log entry:

- `insertId`: A unique, auto-generated ID for the log entry.
- `logName`: The name of the log entry. Common values include:
  - `projects/<project_id>/logs/kubelet` for kubelet logs
  - `projects/<project_id>/logs/container-runtime` for containerd/Docker logs
  - `projects/<project_id>/logs/node-problem-detector` for node problem detector logs
  - `projects/<project_id>/logs/kube-node-installation` for node installation logs
  - `projects/<project_id>/logs/kube-node-configuration` for node configuration logs
- `receiveTimestamp`: The timestamp that the log entry was received by the logging system.
- `resource`: The monitored resource that the log entry is associated with.
  - `type`: The type of the Monitored Resource. For node logs, this is `k8s_node`.
  - `labels`:
    - `cluster_name`: The name of the Kubernetes cluster.
    - `project_id`: The ID of the GCP project where the GKE cluster is located.
    - `location`: The location of the GKE cluster (region or zone).
    - `node_name`: The name of the GKE node.
- `jsonPayload` or `textPayload`: The payload of the log entry containing the log message.
- `timestamp`: The timestamp of when the log entry was emitted.
- `severity`: The severity level of the log entry (DEBUG, INFO, WARNING, ERROR, etc.).

## Sample Queries

### List all kubelet logs for a cluster

This query lists all kubelet logs for a given cluster, project, and location.

```lql
resource.type="k8s_node"
logName="projects/<project_id>/logs/kubelet"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```

### List kubelet logs for a specific node

This query lists kubelet logs for a specific node in the cluster.

```lql
resource.type="k8s_node"
logName="projects/<project_id>/logs/kubelet"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
resource.labels.node_name="<node_name>"
```

### List container runtime logs

This query lists all container runtime (containerd/Docker) logs for a cluster.

```lql
resource.type="k8s_node"
logName="projects/<project_id>/logs/container-runtime"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```

### List node problem detector logs

This query lists logs from the node problem detector, useful for identifying hardware
and kernel-level issues.

```lql
resource.type="k8s_node"
logName="projects/<project_id>/logs/node-problem-detector"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
```

### List error-level node logs

This query lists all node logs with ERROR severity or higher across all system components.

```lql
resource.type="k8s_node"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
severity>=ERROR
```

### List all system component logs for a node

This query lists logs from all system components for a specific node.

```lql
resource.type="k8s_node"
resource.labels.cluster_name="<cluster_name>"
resource.labels.location="<location>"
resource.labels.project_id="<project_id>"
resource.labels.node_name="<node_name>"
```
