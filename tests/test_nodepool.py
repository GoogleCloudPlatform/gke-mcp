import pytest
import json
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.nodepool import (
    create_node_pool,
    list_node_pools,
    get_node_pool,
    update_node_pool,
    delete_node_pool
)

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    cfg.enable_delete_tools = True
    return cfg

@patch("gke_mcp.tools.nodepool._get_client")
def test_create_node_pool(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.create_node_pool.return_value = mock_response
    
    node_pool_json = '{"name": "pool1", "initialNodeCount": 3}'
    
    with patch("gke_mcp.tools.nodepool.MessageToDict", return_value={"name": "pool1"}):
        res = create_node_pool(mock_config, "proj1", "us-central1", "my-cluster", node_pool_json)
        assert "pool1" in res
        mock_client.create_node_pool.assert_called_once_with(
            parent="projects/proj1/locations/us-central1/clusters/my-cluster",
            node_pool={"name": "pool1", "initialNodeCount": 3}
        )

@patch("gke_mcp.tools.nodepool._get_client")
def test_list_node_pools(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.list_node_pools.return_value = mock_response
    
    with patch("gke_mcp.tools.nodepool.MessageToDict", return_value={"nodePools": []}):
        res = list_node_pools(mock_config, "proj1", "us-central1", "my-cluster")
        assert "nodePools" in res
        mock_client.list_node_pools.assert_called_once_with(
            parent="projects/proj1/locations/us-central1/clusters/my-cluster"
        )

@patch("gke_mcp.tools.nodepool._get_client")
def test_get_node_pool(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.get_node_pool.return_value = mock_response
    
    with patch("gke_mcp.tools.nodepool.MessageToDict", return_value={"name": "pool1"}):
        res = get_node_pool(mock_config, "proj1", "us-central1", "my-cluster", "pool1")
        assert "pool1" in res
        mock_client.get_node_pool.assert_called_once_with(
            name="projects/proj1/locations/us-central1/clusters/my-cluster/nodePools/pool1"
        )

@patch("gke_mcp.tools.nodepool._get_client")
def test_update_node_pool(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.update_node_pool.return_value = mock_response
    
    update_json = '{"initialNodeCount": 5}'
    
    with patch("gke_mcp.tools.nodepool.MessageToDict", return_value={"name": "pool1"}):
        res = update_node_pool(mock_config, "proj1", "us-central1", "my-cluster", "pool1", update_json)
        assert "pool1" in res
        mock_client.update_node_pool.assert_called_once_with(
            request={
                "initialNodeCount": 5,
                "name": "projects/proj1/locations/us-central1/clusters/my-cluster/nodePools/pool1"
            }
        )

@patch("gke_mcp.tools.nodepool._get_client")
def test_delete_node_pool(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.delete_node_pool.return_value = mock_response
    
    with patch("gke_mcp.tools.nodepool.MessageToDict", return_value={"status": "DELETING"}):
        res = delete_node_pool(mock_config, "proj1", "us-central1", "my-cluster", "pool1")
        assert "DELETING" in res
        mock_client.delete_node_pool.assert_called_once_with(
            name="projects/proj1/locations/us-central1/clusters/my-cluster/nodePools/pool1"
        )
