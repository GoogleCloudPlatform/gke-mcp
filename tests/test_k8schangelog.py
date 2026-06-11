import pytest
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.k8schangelog import keep_only_changes, get_k8s_changelog

def test_keep_only_changes():
    changelog_data = """
# Changelog
Some introductory text.

# v1.33.0
Changelog details for v1.33.0.

## Downloads for v1.33.0
Download URLs here.

## Changelog since v1.32.0
Detailed notes.

## Dependencies
- dependency A
- dependency B

# v1.33.1
Changelog details for v1.33.1.
"""
    res = keep_only_changes(changelog_data)
    
    assert "v1.33.0" in res
    assert "Changelog details for v1.33.0." in res
    assert "Detailed notes." in res
    assert "v1.33.1" in res
    
    # Excluded sections
    assert "Downloads for" not in res
    assert "Download URLs here." not in res
    assert "Dependencies" not in res
    assert "dependency A" not in res

@patch("gke_mcp.tools.k8schangelog.urllib.request.urlopen")
def test_get_k8s_changelog(mock_urlopen):
    # Mock HTTP response
    mock_resp = MagicMock()
    mock_resp.status = 200
    mock_resp.read.return_value = b"# v1.33.0\nChangelog details."
    mock_urlopen.return_value.__enter__.return_value = mock_resp
    
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    
    res = get_k8s_changelog(cfg, "1.33")
    assert "v1.33.0" in res
    assert "Changelog details." in res
    
    with pytest.raises(ValueError):
        get_k8s_changelog(cfg, "invalid-version")
