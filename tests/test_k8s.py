import pytest
import yaml
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.k8s import (
    _get_context_name,
    _evaluate_jsonpath,
    get_k8s_version,
    get_k8s_resource,
    apply_k8s_manifest,
    list_k8s_events
)

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    return cfg

def test_get_context_name():
    assert _get_context_name("projects/p1/locations/l1/clusters/c1") == "gke_p1_l1_c1"
    with pytest.raises(ValueError):
        _get_context_name("invalid/path")

def test_evaluate_jsonpath():
    obj = {
        "metadata": {"name": "pod1"},
        "spec": {
            "containers": [
                {"image": "nginx"},
                {"image": "redis"}
            ]
        }
    }
    assert _evaluate_jsonpath(obj, "metadata.name") == "pod1"
    assert _evaluate_jsonpath(obj, "spec.containers[0].image") == "nginx"
    assert _evaluate_jsonpath(obj, "spec.containers[1].image") == "redis"
    assert _evaluate_jsonpath(obj, "spec.containers[2].image") == "<none>"
    assert _evaluate_jsonpath(obj, "invalid.path") == "<none>"

@patch("gke_mcp.tools.k8s._get_api_client")
def test_get_k8s_version(mock_get_api_client, mock_config):
    mock_api = MagicMock()
    mock_get_api_client.return_value = mock_api
    
    with patch("kubernetes.client.VersionApi") as mock_version_api_cls:
        mock_api_instance = MagicMock()
        mock_version_api_cls.return_value = mock_api_instance
        
        mock_info = MagicMock()
        mock_info.git_version = "v1.28.3"
        mock_api_instance.get_code.return_value = mock_info
        
        res = get_k8s_version(mock_config, "p1", "l1", "c1")
        assert "Server Version: v1.28.3" in res
        mock_version_api_cls.assert_called_once_with(api_client=mock_api)

@patch("gke_mcp.tools.k8s._get_api_client")
def test_get_k8s_resource_yaml(mock_get_api_client, mock_config):
    mock_api = MagicMock()
    mock_get_api_client.return_value = mock_api
    
    # Mock GAPIC DynamicClient call
    with patch("gke_mcp.tools.k8s.DynamicClient") as mock_dyn_cls:
        mock_dyn = MagicMock()
        mock_dyn_cls.return_value = mock_dyn
        
        mock_resource = MagicMock()
        mock_resource.namespaced = True
        mock_resource.urls = {"namespaced": "/api/v1/namespaces/{namespace}/pods"}
        mock_dyn.resources.get.return_value = mock_resource
        
        mock_api.call_api.return_value = ({"metadata": {"name": "pod1"}, "kind": "Pod", "apiVersion": "v1"}, 200, {})
        
        res = get_k8s_resource(
            mock_config, "p1", "l1", "c1", "Pod", name="pod1", namespace="default", output_format="yaml"
        )
        
        assert "kind: Pod" in res
        assert "name: pod1" in res
        mock_api.call_api.assert_called_once_with(
            resource_path="/api/v1/namespaces/default/pods/pod1",
            method="GET",
            header_params={"Accept": "*/*"},
            auth_settings=["BearerToken"],
            response_type="object"
        )

@patch("gke_mcp.tools.k8s._get_api_client")
def test_apply_k8s_manifest(mock_get_api_client, mock_config):
    mock_api = MagicMock()
    mock_get_api_client.return_value = mock_api
    
    with patch("gke_mcp.tools.k8s.DynamicClient") as mock_dyn_cls:
        mock_dyn = MagicMock()
        mock_dyn_cls.return_value = mock_dyn
        
        mock_resource = MagicMock()
        mock_resource.namespaced = True
        mock_dyn.resources.get.return_value = mock_resource
        
        mock_applied = MagicMock()
        mock_applied.to_dict.return_value = {"metadata": {"name": "my-dep"}, "kind": "Deployment"}
        mock_resource.server_side_apply.return_value = mock_applied
        
        manifest = """
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-dep
  namespace: default
"""
        res = apply_k8s_manifest(mock_config, "p1", "l1", "c1", manifest)
        assert "kind: Deployment" in res
        assert "name: my-dep" in res
        mock_resource.server_side_apply.assert_called_once_with(
            body={"apiVersion": "apps/v1", "kind": "Deployment", "metadata": {"name": "my-dep", "namespace": "default"}},
            name="my-dep",
            namespace="default",
            field_manager="gke-mcp-agent",
            force=False,
            dry_run=None
        )
