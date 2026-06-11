import os
import json
import re
import yaml
import logging
from typing import Optional, List, Dict, Any, Tuple
from kubernetes import client, config
from kubernetes.dynamic import DynamicClient
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.k8s")

def _get_context_name(cluster_path: str) -> str:
    # clusterPath format: projects/PROJECT/locations/LOCATION/clusters/CLUSTER
    parts = cluster_path.split("/")
    if len(parts) != 6:
        raise ValueError(f"invalid cluster path: {cluster_path}")
    project = parts[1]
    location = parts[3]
    cluster_name = parts[5]
    return f"gke_{project}_{location}_{cluster_name}"

def _get_api_client(cluster_path: str) -> client.ApiClient:
    context_name = _get_context_name(cluster_path)
    # Load kube config with context override
    return config.new_client_from_config(context=context_name)

def _evaluate_jsonpath(obj: dict, path: str) -> str:
    clean_path = path.strip().lstrip("{").rstrip("}").lstrip(".")
    parts = clean_path.split(".")
    
    current = obj
    for part in parts:
        if not isinstance(current, dict):
            return "<none>"
            
        if "[" in part and part.endswith("]"):
            key, idx_str = part.split("[", 1)
            idx = int(idx_str.rstrip("]"))
            current = current.get(key)
            if isinstance(current, list) and len(current) > idx:
                current = current[idx]
            else:
                return "<none>"
        else:
            current = current.get(part)
            
        if current is None:
            return "<none>"
            
    return str(current)

def _format_custom_columns(items: List[dict], custom_columns: str) -> str:
    if ".." in custom_columns or "?(" in custom_columns:
        raise ValueError("invalid custom column format: recursive descent '..' and filter '?()' expressions are not supported")
        
    columns = custom_columns.split(",")
    headers = []
    paths = []
    
    for col in columns:
        parts = col.split(":", 1)
        if len(parts) != 2:
            raise ValueError(f"invalid custom column format: {col}. Expected HEADER:JSONPATH")
        headers.append(parts[0])
        paths.append(parts[1])
        
    rows = []
    for item in items:
        row = []
        for path in paths:
            row.append(_evaluate_jsonpath(item, path))
        rows.append(row)
        
    if not headers:
        return ""
        
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if i < len(col_widths):
                col_widths[i] = max(col_widths[i], len(cell))
                
    result = []
    result.append("  ".join(h.ljust(col_widths[i]) for i, h in enumerate(headers)))
    for row in rows:
        result.append("  ".join(cell.ljust(col_widths[i]) for i, cell in enumerate(row)))
        
    return "\n".join(result)

def _format_table(table_dict: dict) -> str:
    cols = table_dict.get("columnDefinitions", [])
    headers = [col.get("name", "") for col in cols]
    
    rows = []
    for row in table_dict.get("rows", []):
        rows.append([str(cell) for cell in row.get("cells", [])])
        
    if not headers:
        return ""
        
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            if i < len(col_widths):
                col_widths[i] = max(col_widths[i], len(cell))
                
    result = []
    result.append("  ".join(h.ljust(col_widths[i]) for i, h in enumerate(headers)))
    for row in rows:
        result.append("  ".join(cell.ljust(col_widths[i]) for i, cell in enumerate(row) if i < len(col_widths)))
        
    return "\n".join(result)

def get_k8s_resource(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    resource_type: str,
    name: Optional[str] = None,
    namespace: Optional[str] = None,
    label_selector: Optional[str] = None,
    field_selector: Optional[str] = None,
    output_format: Optional[str] = "table",
    custom_columns: Optional[str] = None
) -> str:
    """Gets one or more Kubernetes resources from a cluster. Resources can be filtered by type, name, namespace, and label selectors. Returns the resources in YAML format. This is similar to running `kubectl get`."""
    cluster_path = f"projects/{project_id}/locations/{location}/clusters/{cluster_name}"
    api_client = _get_api_client(cluster_path)
    dyn_client = DynamicClient(api_client)
    
    # Resolve GVR using DynamicClient resources registry
    try:
        # get matching resource helper
        # Try getting by resource_name (plural like 'pods')
        resource = dyn_client.resources.get(resource_name=resource_type.lower())
    except Exception:
        try:
            # Fallback to kind
            resource = dyn_client.resources.get(kind=resource_type)
        except Exception as e:
            raise ValueError(f"failed to resolve resource type {resource_type}: {e}")
            
    is_namespaced = resource.namespaced
    
    # Choose output transport Accept headers
    output_format = output_format.lower() if output_format else "table"
    use_table = output_format in ("table", "wide") and not custom_columns
    
    # Build URL path
    if is_namespaced and namespace:
        url_base = resource.urls.get("namespaced", "").format(namespace=namespace)
    else:
        # If namespaced but no namespace provided, and we list, we fetch across all namespaces
        if is_namespaced and not name:
            url_base = resource.urls.get("allNamespaces", resource.urls.get("base", ""))
        else:
            url_base = resource.urls.get("base", "")
            
    if name:
        url = f"{url_base}/{name}"
    else:
        url = url_base
        
    # Append selector queries
    queries = []
    if label_selector:
        queries.append(f"labelSelector={label_selector}")
    if field_selector:
        queries.append(f"fieldSelector={field_selector}")
    if queries:
        url = f"{url}?{'&'.join(queries)}"
        
    # Prepare HTTP headers
    accept_header = "*/*"
    if use_table:
        accept_header = "application/json;as=Table;v=v1;g=meta.k8s.io"
        if output_format == "wide":
            accept_header += ",application/json;as=Table;v=v1beta1;g=meta.k8s.io;includeColumns=wide"
            
    try:
        # Call API directly
        # header_params is dict, response_type='object' returns parsed json dict
        response = api_client.call_api(
            resource_path=url,
            method='GET',
            header_params={'Accept': accept_header},
            auth_settings=['BearerToken'], # standard bearer token auth
            response_type='object'
        )
        # response is a tuple: (data, status_code, headers)
        data = response[0]
    except Exception as e:
        raise RuntimeError(f"failed to get resource from api-server: {e}")
        
    if custom_columns:
        if name:
            return _format_custom_columns([data], custom_columns)
        else:
            items = data.get("items", [])
            return _format_custom_columns(items, custom_columns)
            
    if use_table:
        return _format_table(data)
        
    if output_format == "json":
        return json.dumps(data, indent=2)
        
    # Default is YAML format
    # If listing, yaml dump of the list structure
    return yaml.safe_dump(data, default_flow_style=False)

def list_k8s_events(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    name: Optional[str] = None,
    namespace: Optional[str] = None,
    resource_type: Optional[str] = None,
    all_namespaces: bool = False,
    limit: int = 500
) -> str:
    """Retrieves events from a Kubernetes cluster. This is similar to running `kubectl events`."""
    cluster_path = f"projects/{project_id}/locations/{location}/clusters/{cluster_name}"
    api_client = _get_api_client(cluster_path)
    core_v1 = client.CoreV1Api(api_client=api_client)
    
    kind = ""
    api_version = ""
    if resource_type:
        dyn_client = DynamicClient(api_client)
        try:
            resource = dyn_client.resources.get(resource_name=resource_type.lower())
        except Exception:
            resource = dyn_client.resources.get(kind=resource_type)
        kind = resource.kind
        api_version = f"{resource.group}/{resource.version}" if resource.group else resource.version
        
    target_namespace = namespace if namespace else "default"
    if all_namespaces:
        target_namespace = ""
        
    limit = limit if limit > 0 else 500
    
    # Build field selectors
    selector_parts = []
    if not all_namespaces and target_namespace:
        selector_parts.append(f"involvedObject.namespace={target_namespace}")
    if kind:
        selector_parts.append(f"involvedObject.kind={kind}")
    if name:
        selector_parts.append(f"involvedObject.name={name}")
    if api_version:
        selector_parts.append(f"involvedObject.apiVersion={api_version}")
        
    field_selector = ",".join(selector_parts) if selector_parts else None
    
    try:
        if all_namespaces:
            events_response = core_v1.list_event_for_all_namespaces(limit=limit, field_selector=field_selector)
        else:
            events_response = core_v1.list_namespaced_event(namespace=target_namespace, limit=limit, field_selector=field_selector)
    except Exception as e:
        raise RuntimeError(f"failed to list events: {e}")
        
    # Sort events by last timestamp or event_time descending (newest first)
    def get_last_seen(e):
        if e.series and e.series.last_observed_time:
            return e.series.last_observed_time
        if e.last_timestamp:
            return e.last_timestamp
        if e.event_time:
            return e.event_time
        return e.first_timestamp
        
    sorted_events = sorted(events_response.items, key=get_last_seen, reverse=True)
    
    # Format intervals humanly
    from datetime import datetime, timezone
    def human_duration(td):
        secs = int(td.total_seconds())
        if secs < 0:
            return "0s"
        if secs < 60:
            return f"{secs}s"
        mins = secs // 60
        if mins < 60:
            return f"{mins}m"
        hours = mins // 60
        if hours < 24:
            return f"{hours}h"
        days = hours // 24
        return f"{days}d"
        
    def get_interval_str(e):
        now = datetime.now(timezone.utc)
        last_seen = get_last_seen(e)
        first_seen = e.event_time if e.event_time else e.first_timestamp
        
        if not last_seen:
            return "<unknown>"
            
        td_last = now - last_seen
        last_str = human_duration(td_last)
        
        if e.series and e.series.count:
            td_first = now - e.series.last_observed_time
            first_str = human_duration(now - first_seen)
            return f"{last_str} (x{e.series.count} over {first_str})"
        elif e.count and e.count > 1:
            first_str = human_duration(now - first_seen)
            return f"{last_str} (x{e.count} over {first_str})"
        return last_str

    headers = []
    if all_namespaces:
        headers.append("NAMESPACE")
    headers.extend(["LAST SEEN", "TYPE", "REASON", "OBJECT", "MESSAGE"])
    
    rows = []
    for e in sorted_events:
        row = []
        if all_namespaces:
            row.append(e.involved_object.namespace or "")
        row.extend([
            get_interval_str(e),
            e.type or "",
            e.reason or "",
            f"{e.involved_object.kind}/{e.involved_object.name}",
            (e.message or "").replace("\n", " ").replace("\t", " ")
        ])
        rows.append(row)
        
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            col_widths[i] = max(col_widths[i], len(cell))
            
    result = []
    result.append("  ".join(h.ljust(col_widths[i]) for i, h in enumerate(headers)))
    for row in rows:
        result.append("  ".join(cell.ljust(col_widths[i]) for i, cell in enumerate(row)))
        
    return "\n".join(result)

def get_k8s_version(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str
) -> str:
    """Retrieves the Kubernetes server version for a given cluster. This is similar to running kubectl version."""
    cluster_path = f"projects/{project_id}/locations/{location}/clusters/{cluster_name}"
    api_client = _get_api_client(cluster_path)
    version_api = client.VersionApi(api_client=api_client)
    
    try:
        version_info = version_api.get_code()
        return f"Server Version: {version_info.git_version}"
    except Exception as e:
        raise RuntimeError(f"failed to get server version: {e}")

def apply_k8s_manifest(
    cfg: Config,
    project_id: str,
    location: str,
    cluster_name: str,
    yaml_manifest: str,
    force_conflicts: bool = False,
    dry_run: bool = False
) -> str:
    """Applies a Kubernetes manifest to a cluster using server-side apply. This is similar to running `kubectl apply --server-side`."""
    cluster_path = f"projects/{project_id}/locations/{location}/clusters/{cluster_name}"
    api_client = _get_api_client(cluster_path)
    dyn_client = DynamicClient(api_client)
    
    try:
        objects = list(yaml.safe_load_all(yaml_manifest))
    except Exception as e:
        raise ValueError(f"failed to parse YAML manifest: {e}")
        
    # Filter empty documents
    objects = [obj for obj in objects if obj is not None]
    if not objects:
        return "No resources found to apply."
        
    applied_docs = []
    errors = []
    
    for i, obj in enumerate(objects):
        kind = obj.get("kind")
        api_version = obj.get("apiVersion")
        name = obj.get("metadata", {}).get("name")
        
        if not kind or not api_version or not name:
            errors.append(f"document {i+1}: missing kind, apiVersion or metadata.name")
            continue
            
        try:
            resource = dyn_client.resources.get(api_version=api_version, kind=kind)
        except Exception as e:
            errors.append(f"document {i+1} ({kind}): failed to resolve resource schema: {e}")
            continue
            
        # Call server-side apply
        dry_run_val = "All" if dry_run else None
        namespace = obj.get("metadata", {}).get("namespace")
        
        if resource.namespaced and not namespace:
            errors.append(f"document {i+1} ({kind}/{name}): namespace is required but not specified")
            continue
            
        try:
            applied_obj = resource.server_side_apply(
                body=obj,
                name=name,
                namespace=namespace if resource.namespaced else None,
                field_manager="gke-mcp-agent",
                force=force_conflicts,
                dry_run=dry_run_val
            )
            # convert applied_obj to dict and serialize to YAML
            applied_yaml = yaml.safe_dump(applied_obj.to_dict(), default_flow_style=False)
            applied_docs.append(f"---\n{applied_yaml}")
        except Exception as e:
            errors.append(f"apply resource {namespace}/{name} (kind: {kind}): {e}")
            
    result = "".join(applied_docs)
    if errors:
        result += "\nErrors:\n" + "\n".join(errors)
        
    if errors:
        raise RuntimeError(result)
        
    return result
