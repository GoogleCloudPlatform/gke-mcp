import pytest
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.clustertoolkit import cluster_toolkit_download

@patch("gke_mcp.tools.clustertoolkit.subprocess.run")
def test_cluster_toolkit_download(mock_run):
    mock_run.return_value = MagicMock(stdout="Cloning...", returncode=0)
    
    cfg = MagicMock(spec=Config)
    res = cluster_toolkit_download(cfg, "/my/path")
    assert "Cloning..." in res
    mock_run.assert_called_once_with(
        ["git", "clone", "https://github.com/GoogleCloudPlatform/cluster-toolkit.git", "/my/path/cluster-toolkit"],
        capture_output=True,
        text=True,
        check=True
    )

def test_cluster_toolkit_download_empty():
    cfg = MagicMock(spec=Config)
    with pytest.raises(ValueError):
        cluster_toolkit_download(cfg, "")
