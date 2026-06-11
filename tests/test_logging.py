import pytest
from datetime import timedelta
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.logging import (
    parse_relative_duration,
    get_log_schema,
    query_logs
)

def test_parse_relative_duration():
    assert parse_relative_duration("10s") == timedelta(seconds=10)
    assert parse_relative_duration("5m") == timedelta(minutes=5)
    assert parse_relative_duration("2h") == timedelta(hours=2)
    with pytest.raises(ValueError):
        parse_relative_duration("invalid")
    with pytest.raises(ValueError):
        parse_relative_duration("5d")

def test_get_log_schema():
    content = get_log_schema("k8s_audit_logs")
    assert "Kubernetes Audit Logs" in content or "audit" in content.lower()
    
    with pytest.raises(ValueError):
        get_log_schema("unsupported_log_type")

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    return cfg

@patch("gke_mcp.tools.logging.cloud_logging.Client")
def test_query_logs_with_format(mock_client_cls, mock_config):
    mock_client = MagicMock()
    mock_client_cls.return_value = mock_client
    
    mock_entry = MagicMock()
    mock_entry.to_api_repr.return_value = {
        "timestamp": "2026-05-13T17:52:58Z",
        "severity": "INFO",
        "textPayload": "test log output"
    }
    mock_client.list_entries.return_value = [mock_entry]
    
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    
    res = query_logs(
        cfg,
        project_id="my-project",
        query="resource.type=gke_cluster",
        limit=5,
        format="{{.timestamp}} [{{.severity}}] {{.textPayload}}"
    )
    
    assert "Result:\n\n2026-05-13T17:52:58Z [INFO] test log output" in res
    mock_client.list_entries.assert_called_once()
