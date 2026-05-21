import json
import logging
from typing import Optional
from google.cloud import monitoring_v3
from google.api_core.gapic_v1.client_info import ClientInfo
from google.protobuf.json_format import MessageToDict
from google.api_core.client_options import ClientOptions
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.monitoring")

def list_monitored_resource_descriptors(cfg: Config, project_id: Optional[str] = None) -> str:
    """List monitored resource descriptors(schema) related to GKE for this project. Prefer to use this tool instead of gcloud"""
    proj = project_id if project_id else cfg.default_project_id
    if not proj:
        raise ValueError("project_id argument cannot be empty")
        
    client_info = ClientInfo(user_agent=cfg.user_agent)
    client = monitoring_v3.MetricServiceClient(client_info=client_info)
    name = f"projects/{proj}"
    
    logger.info(f"Listing monitored resource descriptors for {name}")
    
    results = []
    try:
        descriptors = client.list_monitored_resource_descriptors(name=name)
        for desc in descriptors:
            desc_dict = MessageToDict(desc._pb)
            results.append(json.dumps(desc_dict, indent=2))
    except Exception as e:
        raise RuntimeError(f"failed to list monitored resource descriptors: {e}")
        
    return "\n".join(results)
