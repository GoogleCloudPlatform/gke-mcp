import requests
import logging
from typing import List, Optional

logger = logging.getLogger("gke-mcp.clients.dk")

class DeveloperKnowledgeClient:
    def get_documents(self, document_ids: List[str]) -> str:
        raise NotImplementedError("get_documents not implemented")

    def answer_query(self, query: str) -> str:
        raise NotImplementedError("answer_query not implemented")

    def search_documents(self, query: str) -> str:
        raise NotImplementedError("search_documents not implemented")

class RealDeveloperKnowledgeClient(DeveloperKnowledgeClient):
    def __init__(self, base_url: str, api_key: str, user_agent: str):
        self.base_url = base_url.rstrip("/")
        self.api_key = api_key
        self.user_agent = user_agent
        self.session = requests.Session()

    def _do_post(self, path: str, payload: dict) -> str:
        url = f"{self.base_url}/{path.lstrip('/')}"
        headers = {
            "Content-Type": "application/json",
            "User-Agent": self.user_agent
        }
        if self.api_key:
            headers["X-Goog-Api-Key"] = self.api_key
            
        logger.info(f"Posting to Developer Knowledge API: {url}")
        
        try:
            resp = self.session.post(url, json=payload, headers=headers, timeout=30)
            if resp.status_code != 200:
                raise RuntimeError(f"API request failed with status {resp.status_code}: {resp.text}")
            return resp.text
        except Exception as e:
            raise RuntimeError(f"Developer Knowledge API request error: {e}")

    def get_documents(self, document_ids: List[str]) -> str:
        raise NotImplementedError("get_documents not implemented")

    def answer_query(self, query: str) -> str:
        return self._do_post("/v1alpha/TopLevel:answerQuery", {"query": query})

    def search_documents(self, query: str) -> str:
        return self._do_post("/v1/documents:searchDocumentChunks", {"query": query})
