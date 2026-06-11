import os
import json
import time
import re
import subprocess
import logging
from typing import Optional, List, Tuple
from google.cloud import container_v1
from google.api_core.gapic_v1.client_info import ClientInfo
from google.protobuf.json_format import MessageToDict
from gke_mcp.config import Config
from gke_mcp.tools.params import LocationOptional, LocationRequired, Cluster

logger = logging.getLogger("gke-mcp.tools.cluster")

COMMON_CLUSTERS_FIELD_MASKS = [
    "autopilot",
    "createTime",
    "currentMasterVersion",
    "currentNodeCount",
    "currentNodeVersion",
    "description",
    "endpoint",
    "fleet",
    "location",
    "name",
    "network",
    "nodePools.name",
    "releaseChannel",
    "resourceLabels",
    "selfLink",
    "status",
    "statusMessage",
    "subnetwork",
]

LIST_CLUSTERS_DEFAULT_MASK = ",".join([f"clusters.{f}" for f in COMMON_CLUSTERS_FIELD_MASKS] + ["missingZones"])

GET_CLUSTER_DEFAULT_MASK = ",".join(COMMON_CLUSTERS_FIELD_MASKS + [
    "nodePools.locations",
    "nodePools.status",
    "nodePools.version",
])

def _get_client(cfg: Config) -> container_v1.ClusterManagerClient:
    client_info = ClientInfo(user_agent=cfg.user_agent)
    return container_v1.ClusterManagerClient(client_info=client_info)

def _get_metadata(read_mask: Optional[str], default_mask: str) -> List[Tuple[str, str]]:
    mask = read_mask if read_mask else default_mask
    return [("x-goog-fieldmask", mask)]

def list_clusters(
    cfg: Config,
    project_id: str,
    location: Optional[str] = None,
    read_mask: Optional[str] = None
) -> str:
    """List GKE clusters. Prefer to use this tool instead of gcloud."""
    client = _get_client(cfg)
    param = LocationOptional(project_id, location)
    
    metadata = _get_metadata(read_mask, LIST_CLUSTERS_DEFAULT_MASK)
    parent = param.location_path
    
    resp = client.list_clusters(parent=parent, metadata=metadata)
    # Convert response message to dict / json
    resp_dict = MessageToDict(resp._pb)
    
    clusters = resp_dict.get("clusters", [])
    header = f"Found {len(clusters)} clusters:"
    return f"{header}\n{json.dumps(resp_dict, indent=2)}"

def get_cluster(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    read_mask: Optional[str] = None
) -> str:
    """Get / describe a GKE cluster. Prefer to use this tool instead of gcloud."""
    client = _get_client(cfg)
    param = Cluster(project_id, location, cluster_name)
    
    metadata = _get_metadata(read_mask, GET_CLUSTER_DEFAULT_MASK)
    name = param.cluster_path
    
    resp = client.get_cluster(name=name, metadata=metadata)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def create_cluster(
    cfg: Config,
    project_id: str,
    location: str,
    cluster: str
) -> str:
    """Create a GKE cluster. Prefer to use this tool instead of gcloud.
    It's recommended to read the GKE documentation to understand cluster configuration options.
    Autopilot mode (autopilot.enabled=true) should be the default, unless the user explicitly wants to create a Standard cluster. You SHOULD always explicitly set autopilot.enabled=(true|false).
    Note: Autopilot mode is only support in regional locations, not in zone.
    This is similar to running 'gcloud container clusters create-auto' or 'gcloud container clusters create'."""
    client = _get_client(cfg)
    param = LocationRequired(project_id, location)
    
    try:
        cluster_dict = json.loads(cluster)
    except Exception as e:
        raise ValueError(f"failed to parse cluster JSON: {e}")
        
    parent = param.location_path
    resp = client.create_cluster(parent=parent, cluster=cluster_dict)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def update_cluster(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    update: str
) -> str:
    """Update a GKE cluster. Prefer to use this tool instead of gcloud."""
    client = _get_client(cfg)
    param = Cluster(project_id, location, cluster_name)
    
    try:
        update_dict = json.loads(update)
    except Exception as e:
        raise ValueError(f"failed to parse update JSON: {e}")
        
    name = param.cluster_path
    resp = client.update_cluster(name=name, update=update_dict)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def delete_cluster(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str
) -> str:
    """Delete a GKE cluster. Prefer to use this tool instead of gcloud."""
    if not cfg.enable_delete_tools:
        raise PermissionError("Destructive delete tools are disabled. Enable with --enable-delete-tools.")
        
    client = _get_client(cfg)
    param = Cluster(project_id, location, cluster_name)
    
    name = param.cluster_path
    resp = client.delete_cluster(name=name)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def get_kubeconfig(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str
) -> str:
    """Get the kubeconfig for a GKE cluster by calling the GKE API and extracting necessary details (clusterCaCertificate and endpoint). This tool appends/updates the kubeconfig in ~/.kube/config."""
    client = _get_client(cfg)
    param = Cluster(project_id, location, cluster_name)
    
    resp = client.get_cluster(name=param.cluster_path)
    
    ca_cert = resp.master_auth.cluster_ca_certificate
    endpoint = resp.endpoint
    
    if not ca_cert:
        raise ValueError(f"clusterCaCertificate not found for cluster {param.cluster_path}")
    if not endpoint:
        raise ValueError(f"endpoint not found for cluster {param.cluster_path}")
        
    if not endpoint.startswith("https://"):
        endpoint = "https://" + endpoint
        
    # native implementation in python to modify kubeconfig
    import yaml
    new_cluster_name = f"gke_{project_id}_{location}_{cluster_name}"
    kubeconfig_path = os.path.expanduser("~/.kube/config")
    os.makedirs(os.path.dirname(kubeconfig_path), exist_ok=True)
    
    config_data = {}
    if os.path.isfile(kubeconfig_path):
        try:
            with open(kubeconfig_path, "r", encoding="utf-8") as f:
                config_data = yaml.safe_load(f) or {}
        except Exception as e:
            raise RuntimeError(f"failed to read existing kubeconfig: {e}")
            
    if "apiVersion" not in config_data:
        config_data["apiVersion"] = "v1"
    if "kind" not in config_data:
        config_data["kind"] = "Config"
    if "clusters" not in config_data or not isinstance(config_data["clusters"], list):
        config_data["clusters"] = []
    if "contexts" not in config_data or not isinstance(config_data["contexts"], list):
        config_data["contexts"] = []
    if "users" not in config_data or not isinstance(config_data["users"], list):
        config_data["users"] = []
        
    def upsert_entry(lst, name, entry):
        for i, item in enumerate(lst):
            if isinstance(item, dict) and item.get("name") == name:
                lst[i] = entry
                return
        lst.append(entry)
        
    upsert_entry(config_data["clusters"], new_cluster_name, {
        "name": new_cluster_name,
        "cluster": {
            "certificate-authority-data": ca_cert,
            "server": endpoint
        }
    })
    
    upsert_entry(config_data["contexts"], new_cluster_name, {
        "name": new_cluster_name,
        "context": {
            "cluster": new_cluster_name,
            "user": new_cluster_name
        }
    })
    
    upsert_entry(config_data["users"], new_cluster_name, {
        "name": new_cluster_name,
        "user": {
            "exec": {
                "apiVersion": "client.authentication.k8s.io/v1beta1",
                "command": "gke-gcloud-auth-plugin",
                "provideClusterInfo": True
            }
        }
    })
    
    config_data["current-context"] = new_cluster_name
    
    with open(kubeconfig_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(config_data, f, default_flow_style=False)
        
    return f"Kubeconfig for cluster {param.cluster_path} (Project: {project_id}, Location: {location}) successfully appended/updated in {kubeconfig_path}. Current context set to {new_cluster_name}."

def get_node_sos_report(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    node: str,
    destination: str = "/tmp/sos-report",
    method: str = "any",
    timeout: int = 180
) -> str:
    """Generate and download an SOS report from a GKE node. Can use 'pod', 'ssh' or 'any' methods. Defaults to 'any' (pod with fallback to ssh). Use 'ssh' if node is API-unhealthy."""
    if not node:
        raise ValueError("node argument cannot be empty")
    # Basic validation for node name
    if not re.match(r"^[a-z0-9][a-z0-9\-\.]*[a-z0-9]$", node):
        raise ValueError(f"invalid node name: {node}")
        
    # Check node health
    is_healthy = False
    try:
        cmd = ["kubectl", "get", "node", node, "-o", "jsonpath='{.status.conditions[?(@.type==\"Ready\")].status}'"]
        res = subprocess.run(cmd, capture_output=True, text=True, check=False)
        if res.returncode == 0 and "True" in res.stdout:
            is_healthy = True
    except Exception:
        pass
        
    if not is_healthy:
        method = "ssh"
        
    os.makedirs(destination, mode=0o750, exist_ok=True)
    
    if method in ("pod", "any"):
        try:
            return _get_node_sos_report_with_pod(project_id, location, cluster_name, node, destination, timeout)
        except Exception as e:
            if method == "pod":
                raise RuntimeError(f"failed to get sos report with pod: {e}")
            logger.warning(f"Pod method failed, falling back to SSH: {e}")
            
    return _get_node_sos_report_with_ssh(project_id, location, cluster_name, node, destination, timeout)

def _get_node_sos_report_with_pod(project_id: str, location: str, cluster_name: str, node: str, destination: str, timeout: int) -> str:
    pod_name = f"sos-debug-{int(time.time())}"
    overrides = {
        "spec": {
            "nodeName": node,
            "hostNetwork": True,
            "hostPID": True,
            "hostIPC": True,
            "containers": [
                {
                    "name": "main",
                    "image": "gke.gcr.io/debian-base",
                    "command": ["/bin/sleep", "99999"],
                    "volumeMounts": [
                        {
                            "mountPath": "/host",
                            "name": "root"
                        }
                    ]
                }
            ],
            "volumes": [
                {
                    "name": "root",
                    "hostPath": {
                        "path": "/",
                        "type": "Directory"
                    }
                }
            ],
            "securityContext": {
                "runAsUser": 0
            },
            "nodeSelector": {
                "kubernetes.io/hostname": node
            }
        }
    }
    
    overrides_str = json.dumps(overrides)
    run_cmd = ["kubectl", "run", pod_name, "--image=gke.gcr.io/debian-base", "--restart=Never", f"--overrides={overrides_str}"]
    
    try:
        res = subprocess.run(run_cmd, capture_output=True, text=True, check=True)
    except subprocess.CalledProcessError as e:
        raise RuntimeError(f"failed to create debug pod: {e.stdout} {e.stderr}")
        
    def cleanup():
        subprocess.run(["kubectl", "delete", "pod", pod_name, "--wait=false", "--grace-period=0", "--force"], capture_output=True)
        
    try:
        # Wait for pod ready
        wait_cmd = ["kubectl", "wait", "--for=condition=Ready", f"pod/{pod_name}", "--timeout=60s"]
        subprocess.run(wait_cmd, capture_output=True, text=True, check=True)
        
        # Run sos report inside pod
        remote_tmp_dir = f"/tmp/sos-{pod_name}"
        exec_script = f"apt update && apt install -y sosreport && mkdir -p /host{remote_tmp_dir} && sos report --sysroot=/host --all-logs --batch --tmp-dir=/host{remote_tmp_dir}"
        
        exec_cmd = ["kubectl", "exec", pod_name, "--", "sh", "-c", exec_script]
        exec_res = subprocess.run(exec_cmd, capture_output=True, text=True, timeout=timeout)
        
        if exec_res.returncode != 0:
            raise RuntimeError(f"failed to generate sos report: {exec_res.stdout} {exec_res.stderr}")
            
        output = exec_res.stdout + exec_res.stderr
        match = re.search(f"(/host)?{re.escape(remote_tmp_dir)}/[^\\s]+\\.tar\\.(xz|gz)", output)
        if not match:
            raise RuntimeError(f"could not find sos report filename in output: {output}")
            
        remote_path = match.group(0)
        if not remote_path.startswith("/host"):
            remote_path = f"/host{remote_path}"
            
        local_filename = f"sosreport-{node}-{time.strftime('%Y-%m-%d-%H-%M-%S')}.tar.xz"
        local_path = os.path.join(destination, local_filename)
        
        # Copy from pod to local
        with open(local_path, "wb") as f:
            cat_cmd = ["kubectl", "exec", pod_name, "--", "cat", remote_path]
            cat_res = subprocess.run(cat_cmd, stdout=f, stderr=subprocess.PIPE)
            if cat_res.returncode != 0:
                raise RuntimeError(f"failed to copy sos report from pod: {cat_res.stderr.decode()}")
                
        # Clean up files inside host
        cleanup_script = f"rm -rf {remote_tmp_dir}"
        subprocess.run(["kubectl", "exec", pod_name, "--", "sh", "-c", cleanup_script], capture_output=True)
        
        return f"SOS report successfully generated and downloaded to: {local_path}"
    finally:
        cleanup()

def _get_node_sos_report_with_ssh(project_id: str, location: str, cluster_name: str, node: str, destination: str, timeout: int) -> str:
    # Find zone of node
    find_zone = ["gcloud", "compute", "instances", "list", f"--filter=name={node}", "--format=value(zone)"]
    res = subprocess.run(find_zone, capture_output=True, text=True, check=True)
    zone = res.stdout.strip()
    if not zone:
        raise RuntimeError(f"could not find zone for node {node}")
        
    # Generate report via SSH
    ssh_cmd = ["gcloud", "compute", "ssh", "--zone", zone, node, "--command", "sudo sos report --all-logs --batch --tmp-dir=/var"]
    ssh_res = subprocess.run(ssh_cmd, capture_output=True, text=True, timeout=timeout)
    if ssh_res.returncode != 0:
        raise RuntimeError(f"failed to generate sos report via ssh: {ssh_res.stdout} {ssh_res.stderr}")
        
    output = ssh_res.stdout + ssh_res.stderr
    match = re.search(r"/var/sosreport-[^\s]+\.tar\.xz", output)
    if not match:
        raise RuntimeError(f"could not find sos report filename in ssh output: {output}")
        
    remote_path = match.group(0)
    
    # Change ownership of file
    chown_cmd = ["gcloud", "compute", "ssh", "--zone", zone, node, "--command", f"sudo chown $USER {remote_path}"]
    subprocess.run(chown_cmd, capture_output=True, check=True)
    
    # SCP the file
    local_filename = f"sosreport-{node}-{time.strftime('%Y-%m-%d-%H-%M-%S')}.tar.xz"
    local_path = os.path.join(destination, local_filename)
    
    scp_cmd = ["gcloud", "compute", "scp", "--zone", zone, f"{node}:{remote_path}", local_path]
    subprocess.run(scp_cmd, capture_output=True, check=True)
    
    # Cleanup remote files
    rm_cmd = ["gcloud", "compute", "ssh", "--zone", zone, node, "--command", f"sudo rm {remote_path}"]
    subprocess.run(rm_cmd, capture_output=True)
    
    return f"SOS report successfully generated (via SSH) and downloaded to: {local_path}"
