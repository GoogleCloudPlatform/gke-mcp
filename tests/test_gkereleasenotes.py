import pytest
from gke_mcp.tools.gkereleasenotes import (
    parse_gke_version,
    compare_versions,
    extract_release_notes_relevant_for_upgrade
)

def test_parse_gke_version():
    assert parse_gke_version("1.33.5-gke.120000") == (1, 33, 5, 120000)
    assert parse_gke_version("1.34.3-gke.240500") == (1, 34, 3, 240500)
    with pytest.raises(ValueError):
        parse_gke_version("1.33")
    with pytest.raises(ValueError):
        parse_gke_version("1.33.5-invalid.12")

def test_compare_versions():
    # Returns 1 if b > a, 0 if b == a, -1 if b < a
    assert compare_versions("1.33.5-gke.120000", "1.34.3-gke.240500") == 1
    assert compare_versions("1.34.3-gke.240500", "1.33.5-gke.120000") == -1
    assert compare_versions("1.33.5-gke.120000", "1.33.5-gke.120000") == 0
    assert compare_versions("1.33.5-gke.120000", "1.33.5-gke.120001") == 1
    assert compare_versions("1.33.5-gke.120002", "1.33.5-gke.120001") == -1

def test_extract_release_notes_relevant_for_upgrade():
    notes = """
    January 1, 2026
    
    1.34.3-gke.240500
    - Fixed a bug in deployments.
    
    December 1, 2025
    
    1.33.5-gke.120000
    - Added node autoprovisioning support.
    
    November 1, 2025
    
    1.30.2-gke.110000
    - Deprecated old API versions.
    """
    
    # We want notes between 1.33.5-gke.120000 and 1.34.3-gke.240500
    res = extract_release_notes_relevant_for_upgrade(notes, "1.33.5-gke.120000", "1.34.3-gke.240500")
    
    assert "1.34.3-gke.240500" in res
    assert "Fixed a bug in deployments" in res
    assert "1.33.5-gke.120000" in res
    assert "Added node autoprovisioning support" in res
    
    # 1.30.2 should be excluded because it is older than sourceVersion
    assert "1.30.2-gke.110000" not in res
    assert "Deprecated old API versions" not in res
