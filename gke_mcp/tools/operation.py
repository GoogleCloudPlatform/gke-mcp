import json
from google.cloud import container_v1
from google.protobuf.json_format import MessageToDict
from gke_mcp.config import Config
from gke_mcp.tools.params import LocationRequired, Operation
from gke_mcp.tools.cluster import _get_client

def list_operations(
    cfg: Config,
    project_id: str,
    location: str
) -> str:
    """List GKE operations in a project and location."""
    client = _get_client(cfg)
    param = LocationRequired(project_id, location)
    
    parent = param.location_path
    resp = client.list_operations(parent=parent)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def get_operation(
    cfg: Config,
    project_id: str,
    location: str,
    operation_id: str
) -> str:
    """Get details of a GKE operation."""
    client = _get_client(cfg)
    param = Operation(project_id, location, operation_id)
    
    name = param.operation_path
    resp = client.get_operation(name=name)
    resp_dict = MessageToDict(resp._pb)
    return json.dumps(resp_dict, indent=2)

def cancel_operation(
    cfg: Config,
    project_id: str,
    location: str,
    operation_id: str
) -> str:
    """Cancel a GKE operation."""
    client = _get_client(cfg)
    param = Operation(project_id, location, operation_id)
    
    name = param.operation_path
    client.cancel_operation(name=name)
    return f"Operation {name} cancelled successfully."
