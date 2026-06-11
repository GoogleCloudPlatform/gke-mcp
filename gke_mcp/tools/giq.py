import logging
from typing import Optional
from google.cloud import gkerecommender_v1
from google.protobuf.json_format import MessageToDict
from google.api_core.gapic_v1.client_info import ClientInfo
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.giq")

def _get_client(cfg: Config) -> gkerecommender_v1.GkeInferenceQuickstartClient:
    client_info = ClientInfo(user_agent=cfg.user_agent)
    return gkerecommender_v1.GkeInferenceQuickstartClient(client_info=client_info)

def generate_inference_manifest(
    cfg: Config,
    model: str,
    model_server: str,
    accelerator: str
) -> str:
    """Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud"""
    if not model:
        raise ValueError("model argument cannot be empty")
    if not model_server:
        raise ValueError("model_server argument cannot be empty")
    if not accelerator:
        raise ValueError("accelerator argument cannot be empty")
        
    client = _get_client(cfg)
    req = {
        "model_server_info": {
            "model": model,
            "model_server": model_server
        },
        "accelerator_type": accelerator
    }
    logger.info(f"Generating optimized manifest with request: {req}")
    
    try:
        resp = client.generate_optimized_manifest(request=req)
        manifests = []
        for m in resp.kubernetes_manifests:
            manifests.append(m.content)
        return "\n---\n".join(manifests)
    except Exception as e:
        raise RuntimeError(f"failed to generate optimized manifest via SDK: {e}")

def fetch_models(cfg: Config) -> str:
    """List all AI models available for GKE via GKE Inference Quickstart (GIQ). Open-source models follow the Huggingface Hub `owner/model_name` format."""
    client = _get_client(cfg)
    logger.info("Fetching available models")
    
    try:
        iterator = client.fetch_models(request={})
        models = list(iterator)
        return "\n".join(models)
    except Exception as e:
        raise RuntimeError(f"failed to fetch models: {e}")

def fetch_model_servers(cfg: Config, model: str) -> str:
    """Fetch available model servers for a given model in GKE Inference Quickstart (GIQ)."""
    if not model:
        raise ValueError("model argument cannot be empty")
        
    client = _get_client(cfg)
    logger.info(f"Fetching model servers for model: {model}")
    
    try:
        iterator = client.fetch_model_servers(request={"model": model})
        servers = list(iterator)
        return "\n".join(servers)
    except Exception as e:
        raise RuntimeError(f"failed to fetch model servers: {e}")

def fetch_profiles(cfg: Config, model: str, model_server: str, model_server_version: str) -> str:
    """Fetch available performance profiles for models and servers in GKE Inference Quickstart (GIQ)."""
    client = _get_client(cfg)
    req = {
        "model": model,
        "model_server": model_server,
        "model_server_version": model_server_version
    }
    logger.info(f"Fetching performance profiles for: {req}")
    
    try:
        iterator = client.fetch_profiles(request=req)
        profiles = []
        for p in iterator:
            # We convert to string representation
            # GAPIC proto messages have a clean str() format or we can format to json dict
            p_dict = MessageToDict(p._pb)
            import json
            profiles.append(json.dumps(p_dict, indent=2))
        return "\n---\n".join(profiles)
    except Exception as e:
        raise RuntimeError(f"failed to fetch profiles: {e}")

def fetch_model_server_versions(cfg: Config, model: str, model_server: str) -> str:
    """Fetch available versions for a given model and model server in GKE Inference Quickstart (GIQ)."""
    if not model:
        raise ValueError("model argument cannot be empty")
    if not model_server:
        raise ValueError("model_server argument cannot be empty")
        
    client = _get_client(cfg)
    req = {
        "model": model,
        "model_server": model_server
    }
    logger.info(f"Fetching model server versions for: {req}")
    
    try:
        iterator = client.fetch_model_server_versions(request=req)
        versions = list(iterator)
        return "\n".join(versions)
    except Exception as e:
        raise RuntimeError(f"failed to fetch model server versions: {e}")
