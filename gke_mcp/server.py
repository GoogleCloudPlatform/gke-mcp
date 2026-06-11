import logging
import sys
from typing import Optional, List
from mcp.server.fastmcp import FastMCP
from gke_mcp.config import Config
from gke_mcp.tools import register_all_tools
from gke_mcp.prompts import register_all_prompts
from gke_mcp.apps.apps import register_all_apps

logger = logging.getLogger("gke-mcp.server")

def check_adc_auth(cfg: Config) -> str:
    from google.cloud import container_v1
    from google.api_core.gapic_v1.client_info import ClientInfo
    
    project_id = cfg.default_project_id
    if not project_id:
        return ""
        
    location = cfg.default_location or "us-central1"
    
    try:
        client_info = ClientInfo(user_agent=cfg.user_agent)
        client = container_v1.ClusterManagerClient(client_info=client_info)
        name = f"projects/{project_id}/locations/{location}"
        # Trigger a test call
        client.get_server_config(name=name)
        return ""
    except Exception as e:
        err_str = str(e).lower()
        if "unauthenticated" in err_str or "credentials" in err_str or "permission" in err_str or "401" in err_str or "403" in err_str:
            warning = (
                "GKE API calls require Application Default Credentials "
                "(https://cloud.google.com/docs/authentication/application-default-credentials). "
                "Get credentials with `gcloud auth application-default login` before calling GKE MCP tools."
            )
            logger.warning(warning)
            return warning
        return ""

def start_server(cfg: Config, server_mode: str, server_host: str, server_port: int, allowed_origins: List[str]):
    # 1. Preflight ADC auth check
    instructions = check_adc_auth(cfg)
    
    # 2. Instantiate FastMCP
    mcp_server = FastMCP(
        name="GKE MCP Server",
        instructions=instructions,
        host=server_host,
        port=server_port
    )
    
    # 3. Register tools, prompts, and apps
    register_all_tools(mcp_server, cfg)
    register_all_prompts(mcp_server)
    register_all_apps(mcp_server, cfg)
    
    # 4. Register Manifest Generation Agent Tool
    from gke_mcp.clients.dk import RealDeveloperKnowledgeClient
    from gke_mcp.agents.manifestgen.agent import Agent
    
    dk_client = RealDeveloperKnowledgeClient(cfg.dk_base_url, cfg.dk_api_key, cfg.user_agent)
    agent = Agent(cfg, dk_client)
    
    @mcp_server.tool(name="generate_manifest")
    def generate_manifest(prompt: str, session_id: Optional[str] = None) -> str:
        """Generates a Kubernetes manifest using Vertex AI based on a description."""
        import uuid
        sess_id = session_id if session_id else str(uuid.uuid4())
        return agent.run(prompt, sess_id)
        
    logger.info(f"Starting GKE MCP Server (Python) in '{server_mode}' transport mode")
    
    # 5. Run transport
    if server_mode == "stdio":
        mcp_server.run(transport="stdio")
    elif server_mode == "http" or server_mode == "sse":
        # FastMCP sse transport starts starlette app using uvicorn
        mcp_server.run(transport="sse")
    else:
        logger.error(f"Unknown transport mode: {server_mode}")
        sys.exit(1)
