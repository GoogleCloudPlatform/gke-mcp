import pytest
import json
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.monitoring import list_monitored_resource_descriptors

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    cfg.default_project_id = "default-proj"
    return cfg

@patch("gke_mcp.tools.monitoring.monitoring_v3.MetricServiceClient")
def test_list_monitored_resource_descriptors(mock_client_cls, mock_config):
    mock_client = MagicMock()
    mock_client_cls.return_value = mock_client
    
    mock_desc = MagicMock()
    mock_desc._pb = MagicMock()
    mock_client.list_monitored_resource_descriptors.return_value = [mock_desc]
    
    with patch("gke_mcp.tools.monitoring.MessageToDict", return_value={"name": "gke_container"}):
        res = list_monitored_resource_descriptors(mock_config)
        assert "gke_container" in res
        mock_client.list_monitored_resource_descriptors.assert_called_once_with(
            name="projects/default-proj"
        )
