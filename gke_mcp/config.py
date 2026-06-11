import logging
import os
import subprocess

logger = logging.getLogger("gke-mcp.config")

class Config:
    def __init__(self, version: str, enable_delete_tools: bool):
        self._user_agent = f"gke-mcp/{version}"
        self._enable_delete_tools = enable_delete_tools

        provider = os.getenv("GKE_MCP_PROVIDER", "")
        if not provider:
            provider = "vertex-ai"
        self._agent_provider = provider

        model = os.getenv("GKE_MCP_MODEL", "")
        if not model:
            if provider == "anthropic":
                model = "claude-opus-4-7"
            else:
                model = "gemini-2.5-pro"
        self._agent_model = model

        self._anthropic_api_key = os.getenv("ANTHROPIC_API_KEY", "")
        self._dk_base_url = os.getenv("DK_BASE_URL", "https://knowledge.googleapis.com")
        self._dk_api_key = os.getenv("DK_API_KEY", "")

        self._default_project_id = self._get_default_project_id()
        self._default_location = self._get_default_location()

    @property
    def user_agent(self) -> str:
        return self._user_agent

    @property
    def default_project_id(self) -> str:
        return self._default_project_id

    @property
    def default_location(self) -> str:
        return self._default_location

    @property
    def agent_provider(self) -> str:
        return self._agent_provider

    @property
    def agent_model(self) -> str:
        return self._agent_model

    @property
    def enable_delete_tools(self) -> bool:
        return self._enable_delete_tools

    @property
    def anthropic_api_key(self) -> str:
        return self._anthropic_api_key

    @property
    def dk_base_url(self) -> str:
        return self._dk_base_url

    @property
    def dk_api_key(self) -> str:
        return self._dk_api_key

    def _get_default_project_id(self) -> str:
        project, err = self._get_gcloud_config("core/project")
        if err:
            logger.warning(f"Failed to get default project ID from gcloud: {err}")
            return ""
        return project

    def _get_default_location(self) -> str:
        region, err = self._get_gcloud_config("compute/region")
        if not err and region:
            return region
        zone, err = self._get_gcloud_config("compute/zone")
        if not err and zone:
            return zone
        return ""

    def _get_gcloud_config(self, key: str) -> tuple[str, str]:
        try:
            # Run gcloud config get <key>
            res = subprocess.run(
                ["gcloud", "config", "get", key],
                capture_output=True,
                text=True,
                check=False
            )
            if res.returncode != 0:
                return "", res.stderr.strip()
            return res.stdout.strip(), ""
        except Exception as e:
            return "", str(e)

def new_test_config(project: str, location: str, provider: str, model: str) -> Config:
    cfg = Config(version="test", enable_delete_tools=False)
    cfg._default_project_id = project
    cfg._default_location = location
    cfg._agent_provider = provider
    cfg._agent_model = model
    return cfg
