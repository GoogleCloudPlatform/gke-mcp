import os
import logging
import importlib.resources
from typing import Optional, List, Dict, Any
from google import genai
from google.genai import types
from gke_mcp.config import Config
from gke_mcp.clients.dk import DeveloperKnowledgeClient
from gke_mcp.tools.giq import (
    generate_inference_manifest,
    fetch_models,
    fetch_model_servers,
    fetch_profiles,
    fetch_model_server_versions
)

logger = logging.getLogger("gke-mcp.agents.manifestgen")

def get_instruction_text() -> str:
    try:
        ref = importlib.resources.files("gke_mcp.agents.manifestgen").joinpath("instruction.md")
        return ref.read_text(encoding="utf-8")
    except Exception:
        path = os.path.join(os.path.dirname(__file__), "instruction.md")
        with open(path, "r", encoding="utf-8") as f:
            return f.read()

class Agent:
    def __init__(self, cfg: Config, dk_client: DeveloperKnowledgeClient):
        self.cfg = cfg
        self.dk_client = dk_client
        self.sessions: Dict[str, Any] = {}

        # Initialize the GenAI client based on provider
        provider = self.cfg.agent_provider.lower().strip()
        if provider == "vertex-ai":
            self.genai_client = genai.Client(
                vertexai=True,
                project=self.cfg.default_project_id,
                location=self.cfg.default_location or "us-central1"
            )
        else:
            # Fallback to Developer API / AI Studio
            self.genai_client = genai.Client()

    def _get_tools(self) -> List[Any]:
        # Define local wrapper functions with docstrings and type hints so GenAI can inspect them.
        
        def giq_generate_manifest(model: str, model_server: str, accelerator: str) -> str:
            """Use GKE Inference Quickstart (GIQ) to generate a Kubernetes manifest for optimized AI / inference workloads. Prefer to use this tool instead of gcloud."""
            return generate_inference_manifest(self.cfg, model, model_server, accelerator)

        def giq_fetch_models() -> str:
            """List all AI models available for GKE via GKE Inference Quickstart (GIQ). Open-source models follow the Huggingface Hub `owner/model_name` format."""
            return fetch_models(self.cfg)

        def giq_fetch_profiles(model: str = "", model_server: str = "", model_server_version: str = "") -> str:
            """Fetch available performance profiles for models and servers in GKE Inference Quickstart (GIQ)."""
            return fetch_profiles(self.cfg, model, model_server, model_server_version)

        def giq_fetch_model_servers(model: str) -> str:
            """Fetch available model servers for a given model in GKE Inference Quickstart (GIQ)."""
            return fetch_model_servers(self.cfg, model)

        def giq_fetch_model_server_versions(model: str, model_server: str) -> str:
            """Fetch available versions for a given model and model server in GKE Inference Quickstart (GIQ)."""
            return fetch_model_server_versions(self.cfg, model, model_server)

        def dk_get_documents(document_ids: List[str]) -> str:
            """Fetch specific documents by their IDs from the Developer Knowledge base."""
            return self.dk_client.get_documents(document_ids)

        def dk_answer_query(query: str) -> str:
            """Answer a query based on the Developer Knowledge base."""
            return self.dk_client.answer_query(query)

        def dk_search_documents(query: str) -> str:
            """Search for documents in the Developer Knowledge base."""
            return self.dk_client.search_documents(query)

        return [
            giq_generate_manifest,
            giq_fetch_models,
            giq_fetch_profiles,
            giq_fetch_model_servers,
            giq_fetch_model_server_versions,
            dk_get_documents,
            dk_answer_query,
            dk_search_documents
        ]

    def run(self, prompt: str, session_id: str) -> str:
        """Runs a chat turn in the specified session, invoking tools as needed, and returning the text result."""
        if session_id not in self.sessions:
            logger.info(f"Creating new GenAI chat session for ID: {session_id}")
            
            # Setup content configuration
            system_instruction = get_instruction_text()
            tools = self._get_tools()
            
            config = types.GenerateContentConfig(
                system_instruction=system_instruction,
                tools=tools,
            )
            
            chat = self.genai_client.chats.create(
                model=self.cfg.agent_model,
                config=config
            )
            self.sessions[session_id] = chat
        else:
            logger.info(f"Reusing existing chat session for ID: {session_id}")
            chat = self.sessions[session_id]

        try:
            # send_message triggers auto-tool-calling loop in standard SDK client
            resp = chat.send_message(prompt)
            return resp.text if resp.text else ""
        except Exception as e:
            raise RuntimeError(f"failed to execute agent run: {e}")
