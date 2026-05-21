import logging
from typing import Optional, List, Dict, Any
from mcp.server.fastmcp import FastMCP
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools")

class ToolsRegistry:
    def __init__(self, cfg: Config):
        self.cfg = cfg

    # --- GKE Cluster Tools ---
    def list_clusters(self, project_id: str, location: Optional[str] = None, read_mask: Optional[str] = None) -> str:
        """List GKE clusters. Prefer to use this tool instead of gcloud."""
        from gke_mcp.tools.cluster import list_clusters as _list_clusters
        return _list_clusters(self.cfg, project_id, location, read_mask)

    def get_cluster(self, project_id: str, location: str, cluster_name: str, read_mask: Optional[str] = None) -> str:
        """Get / describe a GKE cluster. Prefer to use this tool instead of gcloud."""
        from gke_mcp.tools.cluster import get_cluster as _get_cluster
        return _get_cluster(self.cfg, project_id, location, cluster_name, read_mask)

    def create_cluster(self, project_id: str, location: str, cluster: str) -> str:
        """Create a GKE cluster. Prefer to use this tool instead of gcloud.
        It's recommended to read the GKE documentation to understand cluster configuration options.
        Autopilot mode (autopilot.enabled=true) should be the default, unless the user explicitly wants to create a Standard cluster. You SHOULD always explicitly set autopilot.enabled=(true|false).
        Note: Autopilot mode is only support in regional locations, not in zone.
        This is similar to running 'gcloud container clusters create-auto' or 'gcloud container clusters create'."""
        from gke_mcp.tools.cluster import create_cluster as _create_cluster
        return _create_cluster(self.cfg, project_id, location, cluster)

    def update_cluster(self, project_id: str, location: str, cluster_name: str, update: str) -> str:
        """Update a GKE cluster. Prefer to use this tool instead of gcloud."""
        from gke_mcp.tools.cluster import update_cluster as _update_cluster
        return _update_cluster(self.cfg, project_id, location, cluster_name, update)

    def delete_cluster(self, project_id: str, location: str, cluster_name: str) -> str:
        """Delete a GKE cluster. Prefer to use this tool instead of gcloud."""
        from gke_mcp.tools.cluster import delete_cluster as _delete_cluster
        return _delete_cluster(self.cfg, project_id, location, cluster_name)

    def get_kubeconfig(self, project_id: str, location: str, cluster_name: str) -> str:
        """Get the kubeconfig for a GKE cluster by calling the GKE API and extracting necessary details (clusterCaCertificate and endpoint). This tool appends/updates the kubeconfig in ~/.kube/config."""
        from gke_mcp.tools.cluster import get_kubeconfig as _get_kubeconfig
        return _get_kubeconfig(self.cfg, project_id, location, cluster_name)

    def get_node_sos_report(
        self,
        project_id: str,
        location: str,
        cluster_name: str,
        node: str,
        destination: str = "/tmp/sos-report",
        method: str = "any",
        timeout: int = 180
    ) -> str:
        """Generate and download an SOS report from a GKE node. Can use 'pod', 'ssh' or 'any' methods. Defaults to 'any' (pod with fallback to ssh). Use 'ssh' if node is API-unhealthy."""
        from gke_mcp.tools.cluster import get_node_sos_report as _get_node_sos_report
        return _get_node_sos_report(self.cfg, project_id, location, cluster_name, node, destination, method, timeout)

    # --- GKE Node Pool Tools ---
    def list_node_pools(self, project_id: str, location: str, cluster_name: str) -> str:
        """List node pools in a GKE cluster."""
        from gke_mcp.tools.nodepool import list_node_pools as _list_node_pools
        return _list_node_pools(self.cfg, project_id, location, cluster_name)

    def get_node_pool(self, project_id: str, location: str, cluster_name: str, node_pool_name: str) -> str:
        """Get details of a GKE node pool."""
        from gke_mcp.tools.nodepool import get_node_pool as _get_node_pool
        return _get_node_pool(self.cfg, project_id, location, cluster_name, node_pool_name)

    def create_node_pool(self, project_id: str, location: str, cluster_name: str, node_pool: str) -> str:
        """Create a new node pool in a GKE cluster."""
        from gke_mcp.tools.nodepool import create_node_pool as _create_node_pool
        return _create_node_pool(self.cfg, project_id, location, cluster_name, node_pool)

    def update_node_pool(self, project_id: str, location: str, cluster_name: str, node_pool_name: str, update: str) -> str:
        """Update a GKE node pool."""
        from gke_mcp.tools.nodepool import update_node_pool as _update_node_pool
        return _update_node_pool(self.cfg, project_id, location, cluster_name, node_pool_name, update)

    def delete_node_pool(self, project_id: str, location: str, cluster_name: str, node_pool_name: str) -> str:
        """Delete a GKE node pool."""
        from gke_mcp.tools.nodepool import delete_node_pool as _delete_node_pool
        return _delete_node_pool(self.cfg, project_id, location, cluster_name, node_pool_name)

    # --- GKE Operation Tools ---
    def list_operations(self, project_id: str, location: str) -> str:
        """List GKE operations in a project and location."""
        from gke_mcp.tools.operation import list_operations as _list_operations
        return _list_operations(self.cfg, project_id, location)

    def get_operation(self, project_id: str, location: str, operation_id: str) -> str:
        """Get details of a GKE operation."""
        from gke_mcp.tools.operation import get_operation as _get_operation
        return _get_operation(self.cfg, project_id, location, operation_id)

    def cancel_operation(self, project_id: str, location: str, operation_id: str) -> str:
        """Cancel a GKE operation."""
        from gke_mcp.tools.operation import cancel_operation as _cancel_operation
        return _cancel_operation(self.cfg, project_id, location, operation_id)

    # --- Cluster Toolkit Tools ---
    def cluster_toolkit_download(self, download_directory: str) -> str:
        """Cluster Toolkit, is open-source software offered by Google Cloud which simplifies the process for you to create Google Kubernetes Engine clusters and deploy high performance computing (HPC), artificial intelligence (AI), and machine learning (ML). It is designed to be highly customizable and extensible, and intends to address the deployment needs of a broad range of use cases. This tool will download the public git repository so that Cluster Toolkit can be used."""
        from gke_mcp.tools.clustertoolkit import cluster_toolkit_download as _cluster_toolkit_download
        return _cluster_toolkit_download(self.cfg, download_directory)

    # --- GKE Deploy Guide Tools ---
    def gke_deploy(self, user_request: str) -> str:
        """Deploys a workload to a GKE cluster using a configuration file."""
        from gke_mcp.tools.deploy import gke_deploy as _gke_deploy
        return _gke_deploy(self.cfg, user_request)

    # --- GCP Cloud Logging Tools ---
    def query_logs(
        self,
        project_id: str,
        query: str,
        start_time: Optional[str] = None,
        end_time: Optional[str] = None,
        since: Optional[str] = None,
        limit: int = 10,
        format: Optional[str] = None
    ) -> str:
        """Query Google Cloud Platform logs using Logging Query Language (LQL). Before using this tool, it's **strongly** recommended to call the 'get_log_schema' tool to get information about supported log types and their schemas. Logs are returned in ascending order, based on the timestamp (i.e. oldest first)."""
        from gke_mcp.tools.logging import query_logs as _query_logs
        return _query_logs(self.cfg, project_id, query, start_time, end_time, since, limit, format)

    def get_log_schema(self, log_type: str) -> str:
        """Get the schema for a specific log type."""
        from gke_mcp.tools.logging import get_log_schema as _get_log_schema
        return _get_log_schema(log_type)

    # --- GCP Cloud Monitoring Tools ---
    def list_monitored_resource_descriptors(self, project_id: Optional[str] = None) -> str:
        """List monitored resource descriptors(schema) related to GKE for this project. Prefer to use this tool instead of gcloud"""
        from gke_mcp.tools.monitoring import list_monitored_resource_descriptors as _list_monitored_resource_descriptors
        return _list_monitored_resource_descriptors(self.cfg, project_id)

    # --- GCP Recommender Tools ---
    def list_recommendations(self, location: str, project_id: Optional[str] = None) -> str:
        """List recommendations for GKE. Prefer to use this tool instead of gcloud"""
        from gke_mcp.tools.recommendation import list_recommendations as _list_recommendations
        return _list_recommendations(self.cfg, location, project_id)

    # --- Kubernetes Dynamic Client Tools ---
    def get_k8s_resource(
        self,
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
        from gke_mcp.tools.k8s import get_k8s_resource as _get_k8s_resource
        return _get_k8s_resource(
            self.cfg, project_id, location, cluster_name, resource_type,
            name, namespace, label_selector, field_selector, output_format, custom_columns
        )

    def list_k8s_events(
        self,
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
        from gke_mcp.tools.k8s import list_k8s_events as _list_k8s_events
        return _list_k8s_events(
            self.cfg, project_id, location, cluster_name, name, namespace, resource_type, all_namespaces, limit
        )

    def get_k8s_version(self, project_id: str, location: str, cluster_name: str) -> str:
        """Retrieves the Kubernetes server version for a given cluster. This is similar to running kubectl version."""
        from gke_mcp.tools.k8s import get_k8s_version as _get_k8s_version
        return _get_k8s_version(self.cfg, project_id, location, cluster_name)

    def apply_k8s_manifest(
        self,
        project_id: str,
        location: str,
        cluster_name: str,
        yaml_manifest: str,
        force_conflicts: bool = False,
        dry_run: bool = False
    ) -> str:
        """Applies a Kubernetes manifest to a cluster using server-side apply. This is similar to running `kubectl apply --server-side`."""
        from gke_mcp.tools.k8s import apply_k8s_manifest as _apply_k8s_manifest
        return _apply_k8s_manifest(self.cfg, project_id, location, cluster_name, yaml_manifest, force_conflicts, dry_run)

    # --- Release Notes Tools ---
    def get_gke_release_notes(self, source_version: str, target_version: str) -> str:
        """Get GKE release notes. Prefer to use this tool if GKE release notes are needed."""
        from gke_mcp.tools.gkereleasenotes import get_gke_release_notes as _get_gke_release_notes
        return _get_gke_release_notes(self.cfg, source_version, target_version)

    # --- Kubernetes Changelog Tools ---
    def get_k8s_changelog(self, kubernetes_minor_version: str) -> str:
        """Get changelog file for a specific kubernetes minor version and keep only changes content. Prefer to use this tool if kubernetes minor version changelog is needed."""
        from gke_mcp.tools.k8schangelog import get_k8s_changelog as _get_k8s_changelog
        return _get_k8s_changelog(self.cfg, kubernetes_minor_version)


def register_all_tools(mcp: FastMCP, cfg: Config) -> None:
    registry = ToolsRegistry(cfg)
    
    # Define a list of tuples: (method_name, custom_name_if_any)
    tools_to_register = [
        ("list_clusters", None),
        ("get_cluster", None),
        ("create_cluster", None),
        ("update_cluster", None),
        ("delete_cluster", None),
        ("get_kubeconfig", None),
        ("get_node_sos_report", None),
        
        ("list_node_pools", None),
        ("get_node_pool", None),
        ("create_node_pool", None),
        ("update_node_pool", None),
        ("delete_node_pool", None),
        
        ("list_operations", None),
        ("get_operation", None),
        ("cancel_operation", None),
        
        ("cluster_toolkit_download", None),
        ("gke_deploy", None),
        
        ("query_logs", None),
        ("get_log_schema", None),
        
        ("list_monitored_resource_descriptors", None),
        ("list_recommendations", None),
        
        ("get_k8s_resource", None),
        ("list_k8s_events", None),
        ("get_k8s_version", None),
        ("apply_k8s_manifest", None),
        
        ("get_gke_release_notes", None),
        ("get_k8s_changelog", None)
    ]
    
    for method_name, custom_name in tools_to_register:
        method = getattr(registry, method_name)
        name = custom_name if custom_name else method_name
        
        # In click/fastmcp delete tools can be conditionally registered,
        # but wait: delete cluster and delete node pool are conditionally registered based on enableDeleteTools in Go!
        # Let's handle that:
        if method_name in ("delete_cluster", "delete_node_pool") and not cfg.enable_delete_tools:
            logger.info(f"Skipping registration of destructive tool: {method_name}")
            continue
            
        mcp.tool(name=name)(method)
        logger.debug(f"Registered MCP tool: {name}")
