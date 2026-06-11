import os
import json
import pytest
import yaml
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.cluster import (
    list_clusters,
    get_cluster,
    create_cluster,
    get_kubeconfig,
    GET_CLUSTER_DEFAULT_MASK
)

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    cfg.enable_delete_tools = False
    return cfg

@patch("gke_mcp.tools.cluster._get_client")
def test_list_clusters(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    # Mock GAPIC list_clusters response
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.list_clusters.return_value = mock_response
    
    with patch("gke_mcp.tools.cluster.MessageToDict", return_value={"clusters": [{"name": "c1"}]}):
        res = list_clusters(mock_config, "proj1", "us-central1")
        
        assert "Found 1 clusters" in res
        mock_client.list_clusters.assert_called_once()
        args, kwargs = mock_client.list_clusters.call_args
        assert kwargs["parent"] == "projects/proj1/locations/us-central1"
        # Check that x-goog-fieldmask was set in metadata
        assert kwargs["metadata"][0][0] == "x-goog-fieldmask"

@patch("gke_mcp.tools.cluster._get_client")
def test_get_cluster(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.get_cluster.return_value = mock_response
    
    with patch("gke_mcp.tools.cluster.MessageToDict", return_value={"name": "my-cluster"}):
        res = get_cluster(mock_config, "proj1", "us-central1", "my-cluster")
        assert "my-cluster" in res
        mock_client.get_cluster.assert_called_once_with(
            name="projects/proj1/locations/us-central1/clusters/my-cluster",
            metadata=[("x-goog-fieldmask", GET_CLUSTER_DEFAULT_MASK)]
        )

@patch("gke_mcp.tools.cluster._get_client")
def test_create_cluster(mock_get_client, mock_config):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    mock_response = MagicMock()
    mock_response._pb = MagicMock()
    mock_client.create_cluster.return_value = mock_response
    
    cluster_json = '{"name": "new-cluster", "autopilot": {"enabled": true}}'
    
    with patch("gke_mcp.tools.cluster.MessageToDict", return_value={"name": "new-cluster"}):
        res = create_cluster(mock_config, "proj1", "us-central1", cluster_json)
        assert "new-cluster" in res
        mock_client.create_cluster.assert_called_once_with(
            parent="projects/proj1/locations/us-central1",
            cluster={"name": "new-cluster", "autopilot": {"enabled": True}}
        )

@patch("gke_mcp.tools.cluster._get_client")
def test_get_kubeconfig(mock_get_client, mock_config, tmp_path):
    mock_client = MagicMock()
    mock_get_client.return_value = mock_client
    
    # Mock get_cluster returns ca certificate and endpoint
    mock_cluster = MagicMock()
    mock_cluster.master_auth.cluster_ca_certificate = "MOCK_CA_CERT_DATA"
    mock_cluster.endpoint = "1.2.3.4"
    mock_client.get_cluster.return_value = mock_cluster
    
    # Patch ~/.kube/config path to a temporary file
    temp_kubeconfig = tmp_path / "config"
    with patch("os.path.expanduser", return_value=str(temp_kubeconfig)):
        res = get_kubeconfig(mock_config, "proj1", "us-central1", "my-cluster")
        
        assert "successfully appended/updated" in res
        assert temp_kubeconfig.exists()
        
        # Verify kubeconfig contents
        with open(temp_kubeconfig, "r") as f:
            data = yaml.safe_load(f)
            
        assert data["apiVersion"] == "v1"
        assert data["kind"] == "Config"
        assert len(data["clusters"]) == 1
        assert data["clusters"][0]["name"] == "gke_proj1_us-central1_my-cluster"
        assert data["clusters"][0]["cluster"]["certificate-authority-data"] == "MOCK_CA_CERT_DATA"
        assert data["clusters"][0]["cluster"]["server"] == "https://1.2.3.4"
        assert data["current-context"] == "gke_proj1_us-central1_my-cluster"
