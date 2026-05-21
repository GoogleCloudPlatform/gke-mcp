import pytest
from unittest.mock import MagicMock
from mcp.server.fastmcp import FastMCP
from gke_mcp.prompts import register_all_prompts

def test_register_all_prompts():
    mock_mcp = MagicMock(spec=FastMCP)
    prompts_dict = {}
    
    # Mock the prompt decorator to store registered functions
    def mock_prompt(name):
        def decorator(func):
            prompts_dict[name] = func
            return func
        return decorator
        
    mock_mcp.prompt = mock_prompt
    register_all_prompts(mock_mcp)
    
    assert "gke:cost" in prompts_dict
    assert "gke:deploy" in prompts_dict
    assert "gke:upgrade-risk-report" in prompts_dict
    assert "gke:upgrades-best-practices-risk-report" in prompts_dict
    
    # Test template rendering
    cost_prompt_func = prompts_dict["gke:cost"]
    cost_res = cost_prompt_func(user_question="how to reduce node count?")
    assert "User Question: how to reduce node count?" in cost_res
    assert "GKE cost and optimization expert" in cost_res
    
    deploy_prompt_func = prompts_dict["gke:deploy"]
    deploy_res = deploy_prompt_func(user_request="deploy redis")
    assert "User Request: deploy redis" in deploy_res
    assert "expert GKE (Google Kubernetes Engine) deployment assistant" in deploy_res
