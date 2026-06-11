import json
from google.cloud import container_v1
from google.protobuf.json_format import MessageToDict
from gke_mcp.config import Config
from gke_mcp.tools.params import Cluster, NodePool
from gke_mcp.tools.cluster import _get_client

def create_node_pool(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    node_pool: str
) -> str:
    """Create a new node pool in a GKE cluster."""
    client = _get_client(cfg)
    param = Cluster(project_id, location, cluster_name)
    
    try:
        nodepool_dict = json.loads(node_pool)
    except Exception as e:
        raise ValueError(f"failed to parse node pool JSON: {e}")
        
    parent = param.cluster_path
    resp = client.create_node_pool(parent=parent, node_pool=nodepool_dict)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def list_node_pools(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str
) -> str:
    """List node pools in a GKE cluster."""
    client = _get_client(cfg)
    param = Cluster(project_id, location, cluster_name)
    
    parent = param.cluster_path
    resp = client.list_node_pools(parent=parent)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def get_node_pool(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    node_pool_name: str
) -> str:
    """Get details of a GKE node pool."""
    client = _get_client(cfg)
    param = NodePool(project_id, location, cluster_name, node_pool_name)
    
    name = param.node_pool_path
    resp = client.get_node_pool(name=name)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def update_node_pool(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    node_pool_name: str,
    update: str
) -> str:
    """Update a GKE node pool."""
    client = _get_client(cfg)
    param = NodePool(project_id, location, cluster_name, node_pool_name)
    
    try:
        update_dict = json.loads(update)
    except Exception as e:
        raise ValueError(f"failed to parse update JSON: {e}")
        
    # The request dict needs 'name' field
    update_dict["name"] = param.node_pool_path
    
    resp = client.update_node_pool(request=update_dict)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def delete_node_pool(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    node_pool_name: str
) -> str:
    """Delete a GKE node pool."""
    if not cfg.enable_delete_tools:
        raise PermissionError("Destructive delete tools are disabled. Enable with --enable-delete-tools.")
        
    client = _get_client(cfg)
    param = NodePool(project_id, location, cluster_name, node_pool_name)
    
    name = param.node_pool_path
    resp = client.delete_node_pool(name=name)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)
