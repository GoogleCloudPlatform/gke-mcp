import pytest
from unittest.mock import MagicMock, patch
from gke_mcp.clients.dk import RealDeveloperKnowledgeClient

@patch("gke_mcp.clients.dk.requests.Session")
def test_dk_answer_query(mock_session_cls):
    mock_session = MagicMock()
    mock_session_cls.return_value = mock_session
    
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.text = '{"answer": "GKE details"}'
    mock_session.post.return_value = mock_resp
    
    client = RealDeveloperKnowledgeClient(
        base_url="https://knowledge.googleapis.com",
        api_key="mock-key",
        user_agent="gke-mcp/test"
    )
    
    res = client.answer_query("how to deploy pod?")
    assert "GKE details" in res
    mock_session.post.assert_called_once_with(
        "https://knowledge.googleapis.com/v1alpha/TopLevel:answerQuery",
        json={"query": "how to deploy pod?"},
        headers={
            "Content-Type": "application/json",
            "User-Agent": "gke-mcp/test",
            "X-Goog-Api-Key": "mock-key"
        },
        timeout=30
    )

@patch("gke_mcp.clients.dk.requests.Session")
def test_dk_search_documents(mock_session_cls):
    mock_session = MagicMock()
    mock_session_cls.return_value = mock_session
    
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.text = '{"chunks": []}'
    mock_session.post.return_value = mock_resp
    
    client = RealDeveloperKnowledgeClient(
        base_url="https://knowledge.googleapis.com",
        api_key="mock-key",
        user_agent="gke-mcp/test"
    )
    
    res = client.search_documents("gke")
    assert "chunks" in res
