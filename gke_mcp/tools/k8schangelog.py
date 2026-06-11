import re
import urllib.request
import logging
from typing import List
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.k8schangelog")

KUBERNETES_MINOR_VERSION_REGEXP = re.compile(r"^\d+\.\d+$")
CHANGELOG_VERSION_LINE_REGEXP = re.compile(r"^# v\d\.\d+\.\d+")
IGNORED_SECTION_PREFIXES = ["## Dependencies", "## Downloads for"]
CHANGELOG_HOST_URL = "https://raw.githubusercontent.com"

def keep_only_changes(changelog: str) -> str:
    result = []
    has_met_first_version_heading = False
    is_in_ignored_section = False
    
    lines = changelog.split("\n")
    for line in lines:
        if not has_met_first_version_heading:
            if CHANGELOG_VERSION_LINE_REGEXP.match(line):
                has_met_first_version_heading = True
            else:
                continue
                
        is_ignored_section_header = False
        for prefix in IGNORED_SECTION_PREFIXES:
            if line.startswith(prefix):
                is_in_ignored_section = True
                is_ignored_section_header = True
                break
                
        if is_ignored_section_header:
            continue
            
        if is_in_ignored_section:
            if line.startswith("# ") or line.startswith("## "):
                is_in_ignored_section = False
                
        if not is_in_ignored_section:
            result.append(line)
            
    return "\n".join(result)

def get_k8s_changelog(cfg: Config, kubernetes_minor_version: str) -> str:
    """Get changelog file for a specific kubernetes minor version and keep only changes content. Prefer to use this tool if kubernetes minor version changelog is needed."""
    version = kubernetes_minor_version.strip()
    if not KUBERNETES_MINOR_VERSION_REGEXP.match(version):
        raise ValueError(f"invalid kubernetes minor version: {version}")
        
    url = f"{CHANGELOG_HOST_URL}/kubernetes/kubernetes/refs/heads/master/CHANGELOG/CHANGELOG-{version}.md"
    logger.info(f"Fetching kubernetes changelog from: {url}")
    
    try:
        req = urllib.request.Request(url, headers={"User-Agent": cfg.user_agent})
        with urllib.request.urlopen(req) as resp:
            if resp.status != 200:
                raise RuntimeError(f"failed to get changelog with status code: {resp.status}")
            body = resp.read().decode("utf-8")
    except Exception as e:
        raise RuntimeError(f"failed to get Kubernetes changelog: {e}")
        
    return keep_only_changes(body)
