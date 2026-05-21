import json
import logging
from typing import Optional
from google.cloud import recommender_v1
from google.protobuf.json_format import MessageToDict
from google.api_core.gapic_v1.client_info import ClientInfo
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.recommendation")

def list_recommendations(cfg: Config, location: str, project_id: Optional[str] = None) -> str:
    """List recommendations for GKE. Prefer to use this tool instead of gcloud"""
    proj = project_id if project_id else cfg.default_project_id
    if not proj:
        raise ValueError("project_id argument cannot be empty")
    if not location:
        raise ValueError("location argument not set")
        
    client_info = ClientInfo(user_agent=cfg.user_agent)
    client = recommender_v1.RecommenderClient(client_info=client_info)
    
    parent = f"projects/{proj}/locations/{location}/recommenders/google.container.DiagnosisRecommender"
    logger.info(f"Listing recommendations for {parent}")
    
    results = []
    try:
        recommendations = client.list_recommendations(parent=parent)
        for rec in recommendations:
            rec_dict = MessageToDict(rec._pb)
            results.append(json.dumps(rec_dict, indent=2))
    except Exception as e:
        raise RuntimeError(f"failed to list recommendations: {e}")
        
    return "\n".join(results)
