import pytest
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.operation import (
    list_operations,
    get_operation,
    cancel_operation
)

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    return cfg

@patch("gke_mcp.tools.operation._get_client")
def test_list_operations(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.list_operations.return_value = mock_response
    
    with patch("gke_mcp.tools.operation.MessageToDict", return_value={"operations": []}):
        res = list_operations(mock_config, "proj1", "us-central1")
        assert "operations" in res
        mock_client.list_operations.assert_called_once_with(
            parent="projects/proj1/locations/us-central1"
        )

@patch("gke_mcp.tools.operation._get_client")
def test_get_operation(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.get_operation.return_value = mock_response
    
    with patch("gke_mcp.tools.operation.MessageToDict", return_value={"name": "op1"}):
        res = get_operation(mock_config, "proj1", "us-central1", "op1")
        assert "op1" in res
        mock_client.get_operation.assert_called_once_with(
            name="projects/proj1/locations/us-central1/operations/op1"
        )

@patch("gke_mcp.tools.operation._get_client")
def test_cancel_operation(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    res = cancel_operation(mock_config, "proj1", "us-central1", "op1")
    assert "cancelled successfully" in res
    mock_client.cancel_operation.assert_called_once_with(
        name="projects/proj1/locations/us-central1/operations/op1"
    )
