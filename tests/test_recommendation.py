import pytest
import json
from unittest.mock import MagicMock, patch
from gke_mcp.config import Config
from gke_mcp.tools.recommendation import list_recommendations

@pytest.fixture
def mock_config():
    cfg = MagicMock(spec=Config)
    cfg.user_agent = "gke-mcp/test"
    cfg.default_project_id = "default-proj"
    return cfg

@patch("gke_mcp.tools.recommendation.recommender_v1.RecommenderClient")
def test_list_recommendations(mock_client_cls, mock_config):
    mock_client = MagicMock()
    mock_client_cls.return_value = mock_client
    
    mock_rec = MagicMock()
    mock_rec._pb = MagicMock()
    mock_client.list_recommendations.return_value = [mock_rec]
    
    with patch("gke_mcp.tools.recommendation.MessageToDict", return_value={"name": "rec-123"}):
        res = list_recommendations(mock_config, location="us-central1")
        assert "rec-123" in res
        mock_client.list_recommendations.assert_called_once_with(
            parent="projects/default-proj/locations/us-central1/recommenders/google.container.DiagnosisRecommender"
        )
