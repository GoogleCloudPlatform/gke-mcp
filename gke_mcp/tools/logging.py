import os
import re
import logging
import importlib.resources
import json
from datetime import datetime, timedelta, timezone
from typing import Optional
from google.cloud import logging as cloud_logging
from jinja2 import Template
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.logging")

SUPPORTED_LOG_TYPES = {
    "k8s_audit_logs": True,
    "k8s_application_logs": True,
    "k8s_event_logs": True
}

def parse_relative_duration(dur: str) -> timedelta:
    match = re.match(r"^(\d+)([smh])$", dur.strip())
    if not match:
        raise ValueError(f"invalid relative duration format: {dur}. Supported units: s, m, h.")
    val = int(match.group(1))
    unit = match.group(2)
    if unit == 's':
        return timedelta(seconds=val)
    elif unit == 'm':
        return timedelta(minutes=val)
    elif unit == 'h':
        return timedelta(hours=val)
    raise ValueError(f"unsupported unit: {unit}")

def query_logs(
    cfg: Config,
    project_id: str,
    query: str,
    start_time: Optional[str] = None,
    end_time: Optional[str] = None,
    since: Optional[str] = None,
    limit: int = 10,
    format: Optional[str] = None
) -> str:
    """Query Google Cloud Platform logs using Logging Query Language (LQL). Before using this tool, it's **strongly** recommended to call the 'get_log_schema' tool to get information about supported log types and their schemas. Logs are returned in ascending order, based on the timestamp (i.e. oldest first)."""
    if not project_id:
        raise ValueError("project_id parameter is required")
        
    if limit > 100:
        raise ValueError("limit parameter cannot be greater than 100")
        
    if since and (start_time or end_time):
        raise ValueError("since parameter cannot be used with start_time or end_time")
        
    # Build filter
    filter_parts = [query] if query else []
    
    start_dt = None
    end_dt = None
    
    if since:
        delta = parse_relative_duration(since)
        start_dt = datetime.now(timezone.utc) - delta
    else:
        if start_time:
            start_dt = datetime.fromisoformat(start_time.replace("Z", "+00:00"))
        if end_time:
            end_dt = datetime.fromisoformat(end_time.replace("Z", "+00:00"))
            
    if start_dt:
        filter_parts.append(f'timestamp >= "{start_dt.strftime("%Y-%m-%dT%H:%M:%SZ")}"')
    if end_dt:
        filter_parts.append(f'timestamp <= "{end_dt.strftime("%Y-%m-%dT%H:%M:%SZ")}"')
        
    full_filter = " AND ".join(filter_parts) if filter_parts else ""
    
    # Initialize client
    # Note: client options can take user_agent, but standard Client uses credentials
    client = cloud_logging.Client(project=project_id)
    
    page_size = limit + 1
    
    logger.info(f"Listing log entries with filter: {full_filter}")
    
    try:
        entries_iter = client.list_entries(
            filter_=full_filter,
            page_size=page_size,
            order_by="timestamp asc"
        )
        # Iterate response
        entries = []
        for entry in entries_iter:
            entries.append(entry)
            if len(entries) >= page_size:
                break
    except Exception as e:
        raise RuntimeError(f"failed to query logs: {e}")
        
    truncated = len(entries) > limit
    if truncated:
        entries = entries[:limit]
        
    log_lines = []
    if not entries:
        log_lines.append("No log entries found.")
    else:
        for entry in entries:
            # Use standard API representation dict
            entry_dict = entry.to_api_repr()
            
            if format:
                # Convert Go template syntax to Jinja2
                jinja_format = format.replace("{{.", "{{")
                try:
                    t = Template(jinja_format)
                    rendered = t.render(**entry_dict)
                    log_lines.append(rendered)
                except Exception as e:
                    raise ValueError(f"failed to format log entry with template {format}: {e}")
            else:
                log_lines.append(json.dumps(entry_dict, indent=2))
                
    result_str = "\n".join(log_lines)
    output = f"Project ID: {project_id}\nLQL Query:\n```\n{full_filter}\n```\nResult:\n\n{result_str}"
    
    if truncated:
        output += f"\n\nWarning: Results truncated. The query returned more than the limit of {limit} log entries. You can use the `limit` parameter to request more entries (up to 100)."
        
    return output

def get_log_schema(log_type: str) -> str:
    """Get the schema for a specific log type."""
    if log_type not in SUPPORTED_LOG_TYPES:
        raise ValueError(f"unsupported log_type: {log_type}")
        
    filename = f"{log_type}.md"
    try:
        ref = importlib.resources.files("gke_mcp.tools.schemas").joinpath(filename)
        return ref.read_text(encoding="utf-8")
    except Exception:
        # Fallback to local file lookup
        path = os.path.join(os.path.dirname(__file__), "schemas", filename)
        with open(path, "r", encoding="utf-8") as f:
            return f.read()
