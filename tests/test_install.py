import os
import json
import pytest
from gke_mcp.install import (
    Options,
    install_gemini_cli_extension,
    install_cursor_mcp_extension,
    install_claude_desktop_extension
)

def test_gemini_cli_extension(tmp_path):
    opts = Options(version="1.0.0", project_only=True, developer_mode=False)
    # Override install_dir to use pytest's tmp_path
    opts.install_dir = str(tmp_path)

    install_gemini_cli_extension(opts)

    ext_dir = tmp_path / ".gemini" / "extensions" / "gke-mcp"
    assert ext_dir.exists()

    manifest_path = ext_dir / "gemini-extension.json"
    assert manifest_path.exists()

    with open(manifest_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    assert data["name"] == "gke-mcp"
    assert data["version"] == "1.0.0"
    assert data["mcpServers"]["gke"]["command"] == opts.exe_path

    gemini_md_path = ext_dir / "GEMINI.md"
    assert gemini_md_path.exists()

def test_cursor_mcp_extension(tmp_path):
    opts = Options(version="0.2.0", project_only=True, developer_mode=False)
    opts.install_dir = str(tmp_path)

    # Pre-create an existing mcp.json file
    mcp_dir = tmp_path / ".cursor"
    mcp_dir.mkdir(parents=True)
    mcp_path = mcp_dir / "mcp.json"
    with open(mcp_path, "w", encoding="utf-8") as f:
        json.dump({"mcpServers": {"other-server": {"command": "echo"}}}, f)

    install_cursor_mcp_extension(opts)

    assert mcp_path.exists()
    with open(mcp_path, "r", encoding="utf-8") as f:
        data = json.load(f)
    assert "other-server" in data["mcpServers"]
    assert "gke-mcp" in data["mcpServers"]
    assert data["mcpServers"]["gke-mcp"]["command"] == opts.exe_path

    rule_path = mcp_dir / "rules" / "gke-mcp.mdc"
    assert rule_path.exists()

def test_claude_desktop_extension(tmp_path, monkeypatch):
    opts = Options(version="1.2.3", project_only=True, developer_mode=False)
    opts.install_dir = str(tmp_path)

    # Mock get_claude_desktop_config_path to point inside tmp_path
    config_file = tmp_path / "claude_desktop_config.json"
    monkeypatch.setattr("gke_mcp.install.install.get_claude_desktop_config_path", lambda: str(config_file))

    install_claude_desktop_extension(opts)

    assert config_file.exists()
    with open(config_file, "r", encoding="utf-8") as f:
        data = json.load(f)
    assert data["mcpServers"]["gke-mcp"]["command"] == opts.exe_path
