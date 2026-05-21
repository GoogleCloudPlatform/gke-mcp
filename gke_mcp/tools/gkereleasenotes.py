import os
import re
import urllib.request
import logging
from datetime import datetime
from typing import Tuple
from bs4 import BeautifulSoup
from gke_mcp.config import Config

logger = logging.getLogger("gke-mcp.tools.gkereleasenotes")

GKE_VERSION_REGEXP = re.compile(r"\d+\.\d+\.\d+-gke\.\d+")
RELEASE_DATE_HEADING_REGEXP = re.compile(r"(^|\n)\s*[A-Za-z]+\s+\d+,\s+\d+\s*(\n|$)")

def parse_gke_version(version: str) -> Tuple[int, int, int, int]:
    parts = version.split("-gke.")
    if len(parts) != 2:
        raise ValueError(f"invalid GKE version format: {version}")
        
    k8s_version_part = parts[0]
    gke_version_part = parts[1]
    
    k8s_parts = k8s_version_part.split(".")
    if len(k8s_parts) != 3:
        raise ValueError(f"invalid Kubernetes version part in GKE version: {k8s_version_part}")
        
    try:
        major = int(k8s_parts[0])
        minor = int(k8s_parts[1])
        patch = int(k8s_parts[2])
        gke_patch = int(gke_version_part)
    except Exception as e:
        raise ValueError(f"failed to parse GKE version values: {e}")
        
    return major, minor, patch, gke_patch

def compare_versions(a: str, b: str) -> int:
    """Returns 1 if b > a, 0 if b == a, -1 if b < a."""
    a_maj, a_min, a_pat, a_gke = parse_gke_version(a)
    b_maj, b_min, b_pat, b_gke = parse_gke_version(b)
    
    if b_maj > a_maj:
        return 1
    elif b_maj < a_maj:
        return -1
        
    if b_min > a_min:
        return 1
    elif b_min < a_min:
        return -1
        
    if b_pat > a_pat:
        return 1
    elif b_pat < a_pat:
        return -1
        
    if b_gke > a_gke:
        return 1
    elif b_gke < a_gke:
        return -1
        
    return 0

def extract_release_notes_relevant_for_upgrade(full_release_notes: str, source_version: str, target_version: str) -> str:
    # Find all occurrences of GKE versions in text
    version_locations = []
    for match in GKE_VERSION_REGEXP.finditer(full_release_notes):
        version_locations.append((match.start(), match.end()))
        
    left_border_loc = None
    right_border_loc = None
    
    if version_locations:
        # Release notes are ordered newest to oldest.
        # Find first version <= targetVersion.
        for idx, loc in enumerate(version_locations):
            version = full_release_notes[loc[0]:loc[1]]
            try:
                cmp = compare_versions(version, target_version)
            except Exception:
                continue
                
            # If target_version >= version
            if cmp == 0:
                left_border_loc = loc
                break
            elif cmp > 0:
                if idx == 0:
                    left_border_loc = loc
                else:
                    left_border_loc = version_locations[idx - 1]
                break
                
        # Find first version >= sourceVersion searching from the end.
        for idx in range(len(version_locations)):
            idx_from_end = len(version_locations) - idx - 1
            loc = version_locations[idx_from_end]
            version = full_release_notes[loc[0]:loc[1]]
            try:
                cmp = compare_versions(version, source_version)
            except Exception:
                continue
                
            if cmp == 0:
                right_border_loc = loc
                break
            elif cmp < 0:
                if idx_from_end == len(version_locations) - 1:
                    right_border_loc = loc
                else:
                    right_border_loc = version_locations[idx_from_end + 1]
                break
                
    left_border = left_border_loc[0] if left_border_loc else 0
    right_border = right_border_loc[1] if right_border_loc else len(full_release_notes)
    
    reduced_release_notes = full_release_notes[left_border:right_border]
    
    left_append = ""
    left_cut = full_release_notes[:left_border]
    if left_cut:
        headings = list(RELEASE_DATE_HEADING_REGEXP.finditer(left_cut))
        if not headings:
            left_append = left_cut
        else:
            last_heading = headings[-1]
            left_append = left_cut[last_heading.start():]
            
    right_append = ""
    right_cut = full_release_notes[right_border:]
    if right_cut:
        headings = list(RELEASE_DATE_HEADING_REGEXP.finditer(right_cut))
        if not headings:
            right_append = right_cut
        else:
            first_heading = headings[0]
            right_cut_append_end = max(0, first_heading.start() - 1)
            right_append = right_cut[:right_cut_append_end]
            
    return left_append + reduced_release_notes + right_append

def get_gke_release_notes(cfg: Config, source_version: str, target_version: str) -> str:
    """Get GKE release notes. Prefer to use this tool if GKE release notes are needed."""
    date_str = datetime.now().strftime("%Y-%m-%d")
    cache_path = f"release-notes-{date_str}.html"
    
    out = b""
    if os.path.isfile(cache_path):
        logger.info(f"Reading GKE release notes from cache: {cache_path}")
        with open(cache_path, "rb") as f:
            out = f.read()
    else:
        logger.info("Fetching GKE release notes from web")
        url = "https://cloud.google.com/kubernetes-engine/docs/release-notes"
        try:
            req = urllib.request.Request(url, headers={'User-Agent': cfg.user_agent})
            with urllib.request.urlopen(req) as resp:
                out = resp.read()
            with open(cache_path, "wb") as f:
                f.write(out)
        except Exception as e:
            raise RuntimeError(f"failed to fetch GKE release notes: {e}")
            
    soup = BeautifulSoup(out, 'html.parser')
    
    # Remove version updates and security updates data-text headers
    for elem in soup.select('[data-text$="Version updates"]'):
        p_p = elem.parent.parent if elem.parent and elem.parent.parent else None
        if p_p:
            p_p.decompose()
            
    for elem in soup.select('[data-text$="Security updates"]'):
        p_p = elem.parent.parent if elem.parent and elem.parent.parent else None
        if p_p:
            p_p.decompose()
            
    releases_texts = []
    for elem in soup.select('.releases'):
        releases_texts.append(elem.get_text())
        
    full_text = "".join(releases_texts)
    
    return extract_release_notes_relevant_for_upgrade(full_text, source_version, target_version)
