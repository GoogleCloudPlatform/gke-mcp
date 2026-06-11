import os
import pytest
from gke_mcp.config import Config, new_test_config

def test_new_config_defaults(monkeypatch):
    monkeypatch.delenv("GKE_MCP_PROVIDER", raising=False)
    monkeypatch.delenv("GKE_MCP_MODEL", raising=False)
    monkeypatch.delenv("ANTHROPIC_API_KEY", raising=False)
    monkeypatch.delenv("DK_BASE_URL", raising=False)
    monkeypatch.delenv("DK_API_KEY", raising=False)

    # Mock _get_default_project_id and _get_default_location to avoid running actual gcloud CLI
    monkeypatch.setattr(Config, "_get_default_project_id", lambda self: "mock-project")
    monkeypatch.setattr(Config, "_get_default_location", lambda self: "us-central1")

    cfg = Config("1.0.0", False)
    assert cfg.user_agent == "gke-mcp/1.0.0"
    assert cfg.agent_provider == "vertex-ai"
    assert cfg.agent_model == "gemini-2.5-pro"
    assert cfg.enable_delete_tools is False
    assert cfg.default_project_id == "mock-project"
    assert cfg.default_location == "us-central1"

def test_new_config_with_env_vars(monkeypatch):
    monkeypatch.setenv("GKE_MCP_PROVIDER", "custom-provider")
    monkeypatch.setenv("GKE_MCP_MODEL", "custom-model")
    monkeypatch.setenv("ANTHROPIC_API_KEY", "test-key")
    monkeypatch.setenv("DK_BASE_URL", "https://test-knowledge.com")
    monkeypatch.setenv("DK_API_KEY", "dk-key")

    monkeypatch.setattr(Config, "_get_default_project_id", lambda self: "env-project")
    monkeypatch.setattr(Config, "_get_default_location", lambda self: "us-east1")

    cfg = Config("0.1.0", True)
    assert cfg.user_agent == "gke-mcp/0.1.0"
    assert cfg.agent_provider == "custom-provider"
    assert cfg.agent_model == "custom-model"
    assert cfg.enable_delete_tools is True
    assert cfg.anthropic_api_key == "test-key"
    assert cfg.dk_base_url == "https://test-knowledge.com"
    assert cfg.dk_api_key == "dk-key"
    assert cfg.default_project_id == "env-project"
    assert cfg.default_location == "us-east1"

def test_new_test_config():
    cfg = new_test_config("test-proj", "test-loc", "vertex-ai", "gemini-2.5-flash")
    assert cfg.user_agent == "gke-mcp/test"
    assert cfg.default_project_id == "test-proj"
    assert cfg.default_location == "test-loc"
    assert cfg.agent_provider == "vertex-ai"
    assert cfg.agent_model == "gemini-2.5-flash"
    assert cfg.enable_delete_tools is False
