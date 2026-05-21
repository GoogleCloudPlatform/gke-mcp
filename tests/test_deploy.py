import pytest
from unittest.mock import MagicMock
from gke_mcp.config import Config
from gke_mcp.tools.deploy import gke_deploy

def test_gke_deploy():
    cfg = MagicMock(spec=Config)
    res = gke_deploy(cfg, "deploy nginx.yaml")
    assert "expert GKE (Google Kubernetes Engine) deployment assistant" in res

def test_gke_deploy_empty():
    cfg = MagicMock(spec=Config)
    with pytest.raises(ValueError):
        gke_deploy(cfg, "   ")
