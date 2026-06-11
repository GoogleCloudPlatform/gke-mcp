import pytest
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.clients.dk import DeveloperKnowledgeClient
from gke_mcp.agents.manifestgen.agent import Agent

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.agent_provider = "vertex-ai"
    cfg.default_project_id = "proj1"
    cfg.default_location = "us-central1"
    cfg.agent_model = "gemini-2.5-pro"
    return cfg

@pytest.fixture
def mock_dk():
    return MagicMock(spec=DeveloperKnowledgeClient)

@patch("gke_mcp.agents.manifestgen.agent.genai.Client")
def test_agent_run_creates_and_reuses_sessions(mock_genai_cls, mock_config, mock_dk):
    mock_client = MagicMock()
    mock_genai_cls.return_value = mock_client
    
    mock_chat = MagicMock()
    mock_client.chats.create.return_value = mock_chat
    
    mock_resp = MagicMock()
    mock_resp.text = "Here is your pod manifest"
    mock_chat.send_message.return_value = mock_resp
    
    agent = Agent(mock_config, mock_dk)
    
    # First run: should create session
    res1 = agent.run("nginx deployment", "sess-1")
    assert res1 == "Here is your pod manifest"
    mock_client.chats.create.assert_called_once()
    mock_chat.send_message.assert_called_once_with("nginx deployment")
    
    # Reset mock calls
    mock_client.chats.create.reset_mock()
    mock_chat.send_message.reset_mock()
    
    # Second run: should reuse session
    res2 = agent.run("add 3 replicas", "sess-1")
    assert res2 == "Here is your pod manifest"
    mock_client.chats.create.assert_not_called()
    mock_chat.send_message.assert_called_once_with("add 3 replicas")
