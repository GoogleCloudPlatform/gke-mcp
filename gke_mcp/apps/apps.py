import os
import json
import logging
import importlib.resources
from typing import Optional, List, Dict, Any
from google.cloud import monitoring_v3
from google.api_core.gapic_v1.client_info import ClientInfo
from mcp.server.fastmcp import FastMCP
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.apps")

DROPDOWN_URI = "ui://dropdown/index.html"
TIME_SERIES_CHART_URI = "ui://monitoring_time_series_chart/index.html"
HTML_MIME_TYPE = "text/html;profile=mcp-app"

def get_dropdown_html() -> str:
    try:
        ref = importlib.resources.files("gke_mcp.apps.dist.apps.dropdown").joinpath("index.html")
        return ref.read_text(encoding="utf-8")
    except Exception:
        path = os.path.join(os.path.dirname(__file__), "dist", "apps", "dropdown", "index.html")
        with open(path, "r", encoding="utf-8") as f:
            return f.read()

def get_timeseries_chart_html() -> str:
    try:
        ref = importlib.resources.files("gke_mcp.apps.dist.apps.timeserieschart").joinpath("index.html")
        return ref.read_text(encoding="utf-8")
    except Exception:
        path = os.path.join(os.path.dirname(__file__), "dist", "apps", "timeserieschart", "index.html")
        with open(path, "r", encoding="utf-8") as f:
            return f.read()

class AppsRegistry:
    def __init__(self, cfg: Config):
        self.cfg = cfg

    # --- Dropdown Tool & Resource ---
    def dropdown(self, options: List[str], title: Optional[str] = None) -> dict:
        """Renders an interactive UI dropdown for the user to select an item from a list.
        Use this tool when you need the user to choose one option from a set of available resources (e.g., clusters, regions, namespaces).
        You MUST provide a valid array of 1 or more options.
        Timing: Call this tool immediately before you need the user's input to proceed. Do not ask the user for clarification in plain text; calling this tool serves as your question to the user.
        After calling this tool, STOP and wait for the user to make a selection via the UI.
        Do NOT list the options in your text response; the UI itself serves as the list and confirmation."""
        if not options:
            raise ValueError("options cannot be empty")
            
        return {
            "status": "PENDING_USER_INPUT",
            "options": options,
            "message": "Present these options to the user. Wait until selection is made"
        }

    # --- Time Series Chart Tools ---
    def monitoring_time_series_chart(
        self,
        query: str,
        project_id: Optional[str] = None,
        title: Optional[str] = None,
        x_legend: Optional[str] = None,
        y_legend: Optional[str] = None
    ) -> str:
        """Interactive tool to display time series data using a React Chart. ALWAYS favor using this tool to query metrics rather than outputting raw values so the user gets a visualization. MUST Call `mql_validator` FIRST to catch syntax issues or metric anomalies before running this tool."""
        proj = project_id if project_id else self.cfg.default_project_id
        if not proj:
            raise ValueError("project_id argument cannot be empty")
        if not query:
            raise ValueError("query argument cannot be empty")
            
        return "Rendered time series data in UI component."

    def query_time_series(self, query: str, project_id: Optional[str] = None) -> dict:
        """Internal app tool. Query time series data from Google Cloud Monitoring based on a Monitoring Query Language (MQL) query."""
        proj = project_id if project_id else self.cfg.default_project_id
        if not proj:
            raise ValueError("project_id argument cannot be empty")
        if not query:
            raise ValueError("query argument cannot be empty")
            
        client_info = ClientInfo(user_agent=self.cfg.user_agent)
        client = monitoring_v3.QueryServiceClient(client_info=client_info)
        name = f"projects/{proj}"
        
        logger.info(f"Querying time series data for {name} with MQL: {query}")
        
        try:
            resp = client.query_time_series(request={"name": name, "query": query})
            series = []
            for i, ts_data in enumerate(resp.time_series_data):
                if i >= 100: # maxSeriesLimit
                    break
                series.append(self._map_timeseries_data(ts_data))
            return {"data": series}
        except Exception as e:
            raise RuntimeError(f"failed to query time series: {e}")

    def _map_timeseries_data(self, ts_data) -> dict:
        label_parts = []
        for lv in ts_data.label_values:
            if lv.string_value:
                label_parts.append(lv.string_value)
        label = " ".join(label_parts)
        
        pts = []
        for p in ts_data.point_data:
            if not p.values:
                continue
            first_val = p.values[0]
            val = getattr(first_val, "double_value", None)
            if val is None:
                val = getattr(first_val, "int64_value", None)
                if val is not None:
                    val = float(val)
                    
            if val is None:
                continue
                
            # end_time timestamp converted to ms
            timestamp_ms = int(p.time_interval.end_time.timestamp() * 1000)
            pts.append({
                "timestamp": timestamp_ms,
                "value": val
            })
            
        return {
            "label": label,
            "points": pts
        }

    def mql_validator(self, query: str, project_id: Optional[str] = None) -> dict:
        """A helper tool to validate Monitoring Query Language (MQL) metric strings. MUST be called immediately before calling `monitoring_time_series_chart` or `query_time_series` to ensure the MQL statement compiles correctly. It fetches 1 page of data to verify syntactical and logical correctness. Returns the original string on success, or an error payload explaining the misconfiguration."""
        proj = project_id if project_id else self.cfg.default_project_id
        if not proj:
            raise ValueError("project_id argument cannot be empty")
        if not query:
            raise ValueError("query argument cannot be empty")
            
        client_info = ClientInfo(user_agent=self.cfg.user_agent)
        client = monitoring_v3.QueryServiceClient(client_info=client_info)
        name = f"projects/{proj}"
        
        logger.info(f"Validating MQL query for {name}: {query}")
        
        status = "VALID"
        err_msg = None
        try:
            # Execute validation check by fetching 1 page
            resp = client.query_time_series(request={"name": name, "query": query, "page_size": 1})
        except Exception as e:
            status = "INVALID"
            err_msg = f"MQL validation failed:\n{e}"
            
        res = {
            "status": status,
            "query": query
        }
        if err_msg:
            res["errorMessage"] = err_msg
            
        return res


def register_all_apps(mcp: FastMCP, cfg: Config) -> None:
    registry = AppsRegistry(cfg)

    # 1. Register Dropdown Tool & Resource
    # Note: custom input schema and metadata can be registered in FastMCP
    # FastMCP decorator supports metadata configuration
    mcp.tool(
        name="dropdown",
        description="Renders an interactive UI dropdown for the user to select an item from a list.\nUse this tool when you need the user to choose one option from a set of available resources (e.g., clusters, regions, namespaces).\nYou MUST provide a valid array of 1 or more options.\n\nTiming: Call this tool immediately before you need the user's input to proceed. Do not ask the user for clarification in plain text; calling this tool serves as your question to the user.\nAfter calling this tool, STOP and wait for the user to make a selection via the UI.\nDo NOT list the options in your text response; the UI itself serves as the list and confirmation."
    )(registry.dropdown)

    # 2. Register time series chart tool
    mcp.tool(
        name="monitoring_time_series_chart",
        description="Interactive tool to display time series data using a React Chart. ALWAYS favor using this tool to query metrics rather than outputting raw values so the user gets a visualization. MUST Call `mql_validator` FIRST to catch syntax issues or metric anomalies before running this tool."
    )(registry.monitoring_time_series_chart)

    # 3. Register query time series tool
    mcp.tool(
        name="query_time_series",
        description="Internal app tool. Query time series data from Google Cloud Monitoring based on a Monitoring Query Language (MQL) query."
    )(registry.query_time_series)

    # 4. Register MQL validator tool
    mcp.tool(
        name="mql_validator",
        description="A helper tool to validate Monitoring Query Language (MQL) metric strings. MUST be called immediately before calling `monitoring_time_series_chart` or `query_time_series` to ensure the MQL statement compiles correctly. It fetches 1 page of data to verify syntactical and logical correctness. Returns the original string on success, or an error payload explaining the misconfiguration."
    )(registry.mql_validator)

    # 5. Register Resources
    @mcp.resource(
        uri=DROPDOWN_URI,
        name="GKE Resource Dropdown UI",
        mime_type=HTML_MIME_TYPE,
        description="Dropdown app UI component"
    )
    def read_dropdown() -> str:
        return get_dropdown_html()

    @mcp.resource(
        uri=TIME_SERIES_CHART_URI,
        name="Time Series Chart UI",
        mime_type=HTML_MIME_TYPE,
        description="Time series chart app UI component"
    )
    def read_timeseries_chart() -> str:
        return get_timeseries_chart_html()
