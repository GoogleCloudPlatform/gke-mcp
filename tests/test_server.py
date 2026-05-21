import pytest
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.server import check_adc_auth, start_server

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.default_project_id = "test-proj"
    cfg.default_location = "us-central1"
    cfg.user_agent = "gke-mcp/test"
    return cfg

@patch("google.cloud.container_v1.ClusterManagerClient")
def test_check_adc_auth_success(mock_client_cls, mock_config):
    mock_client = MagicMock()
    mock_client_cls.return_value = mock_client
    
    # Mock successful server config lookup
    mock_client.get_server_config.return_value = MagicMock()
    
    res = check_adc_auth(mock_config)
    assert res == ""
    mock_client.get_server_config.assert_called_once_with(
        name="projects/test-proj/locations/us-central1"
    )

@patch("google.cloud.container_v1.ClusterManagerClient")
def test_check_adc_auth_unauthenticated(mock_client_cls, mock_config):
    mock_client = MagicMock()
    mock_client_cls.return_value = mock_client
    mock_client.get_server_config.side_effect = Exception("401 Unauthenticated: Request lacks valid authentication credentials.")
    
    res = check_adc_auth(mock_config)
    assert "require Application Default Credentials" in res

@patch("gke_mcp.server.FastMCP")
@patch("gke_mcp.server.register_all_tools")
@patch("gke_mcp.server.register_all_prompts")
@patch("gke_mcp.server.register_all_apps")
def test_start_server(mock_reg_apps, mock_reg_prompts, mock_reg_tools, mock_fastmcp_cls, mock_config):
    mock_mcp = MagicMock()
    mock_fastmcp_cls.return_value = mock_mcp
    
    with patch("gke_mcp.server.check_adc_auth", return_value=""):
        start_server(mock_config, "stdio", "127.0.0.1", 8080, ["http://localhost"])
        
        mock_fastmcp_cls.assert_called_once_with(
            name="GKE MCP Server",
            instructions="",
            host="127.0.0.1",
            port=8080
        )
        mock_reg_tools.assert_called_once_with(mock_mcp, mock_config)
        mock_reg_prompts.assert_called_once_with(mock_mcp)
        mock_reg_apps.assert_called_once_with(mock_mcp, mock_config)
        mock_mcp.run.assert_called_once_with(transport="stdio")
