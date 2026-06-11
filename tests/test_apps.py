import pytest
from unittest.mock import MagicMock, patch
from mcp.server.fastmcp import FastMCP
from gke_mcp.config import Config
from gke_mcp.apps import register_all_apps

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.default_project_id = "proj1"
    cfg.user_agent = "gke-mcp/test"
    return cfg

def test_register_all_apps(mock_config):
    mock_mcp = MagicMock(spec=FastMCP)
    tools = {}
    resources = {}
    
    def mock_tool(name, description=None):
        def decorator(func):
            tools[name] = func
            return func
        return decorator
        
    def mock_resource(uri, name, mime_type=None, description=None):
        def decorator(func):
            resources[uri] = func
            return func
        return decorator
        
    mock_mcp.tool = mock_tool
    mock_mcp.resource = mock_resource
    
    register_all_apps(mock_mcp, mock_config)
    
    # Check tool registration
    assert "dropdown" in tools
    assert "monitoring_time_series_chart" in tools
    assert "query_time_series" in tools
    assert "mql_validator" in tools
    
    # Check resource registration
    assert "ui://dropdown/index.html" in resources
    assert "ui://monitoring_time_series_chart/index.html" in resources
    
    # Test dropdown tool return payload
    dropdown_func = tools["dropdown"]
    res = dropdown_func(options=["opt1", "opt2"], title="Choose")
    assert res["status"] == "PENDING_USER_INPUT"
    assert res["options"] == ["opt1", "opt2"]
    
    # Test resource getters read content successfully
    with patch("gke_mcp.apps.apps.get_dropdown_html", return_value="<html>Dropdown</html>"):
        dropdown_html_func = resources["ui://dropdown/index.html"]
        assert dropdown_html_func() == "<html>Dropdown</html>"
