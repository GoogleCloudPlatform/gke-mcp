import pytest
import json
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.giq import (
    generate_inference_manifest,
    fetch_models,
    fetch_model_servers,
    fetch_profiles,
    fetch_model_server_versions
)

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    return cfg

@patch("gke_mcp.tools.giq._get_client")
def test_generate_inference_manifest(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_resp = MagicMock()
    mock_manifest = MagicMock()
    mock_manifest.content = "apiVersion: v1\nkind: Pod"
    mock_resp.kubernetes_manifests = [mock_manifest]
    mock_client.generate_optimized_manifest.return_value = mock_resp
    
    res = generate_inference_manifest(mock_config, "gemma", "vllm", "l4")
    assert "kind: Pod" in res
    mock_client.generate_optimized_manifest.assert_called_once_with(
        request={
            "model_server_info": {"model": "gemma", "model_server": "vllm"},
            "accelerator_type": "l4"
        }
    )

@patch("gke_mcp.tools.giq._get_client")
def test_fetch_models(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    mock_client.fetch_models.return_value = ["model1", "model2"]
    
    res = fetch_models(mock_config)
    assert "model1\nmodel2" in res
    mock_client.fetch_models.assert_called_once_with(request={})
