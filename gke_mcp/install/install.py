import os
import sys
import json
import logging
import importlib.resources
import subprocess
from typing import Dict, Any

logger = logging.getLogger("gke-mcp.install")

CURSOR_RULE_HEADER = """---
name: GKE MCP Instructions
description: Provides guidance for using the gke-mcp tool with Cursor.
alwaysApply: true
---

# GKE MCP Tool Instructions

This rule provides context for using the gke-mcp tool within Cursor.

"""

class Options:
    def __init__(self, version: str, project_only: bool, developer_mode: bool):
        self.version = version
        self.developer_mode = developer_mode

        if project_only:
            self.install_dir = os.getcwd()
        else:
            self.install_dir = os.path.expanduser("~")

        # Resolve executable path.
        # If running from a venv, `gke-mcp` CLI script will be in the same folder as Python interpreter.
        bin_dir = os.path.dirname(sys.executable)
        gke_mcp_bin = os.path.join(bin_dir, "gke-mcp")
        if os.path.isfile(gke_mcp_bin):
            self.exe_path = gke_mcp_bin
        else:
            self.exe_path = "gke-mcp"

def get_gemini_markdown() -> str:
    try:
        ref = importlib.resources.files("gke_mcp.install").joinpath("GEMINI.md")
        return ref.read_text(encoding="utf-8")
    except Exception:
        path = os.path.join(os.path.dirname(__file__), "GEMINI.md")
        with open(path, "r", encoding="utf-8") as f:
            return f.read()

def install_gemini_cli_extension(opts: Options) -> None:
    context_filename = "GEMINI.md"
    if opts.developer_mode:
        # In developer mode, try to use the GEMINI.md in the repository
        # If opts.exe_path is a venv binary, let's find the source path of GEMINI.md
        # Go code did: filepath.Join(filepath.Dir(opts.exePath), "pkg", "install", "GEMINI.md")
        # In python, let's check if we can resolve it from the workspace root.
        repo_root = os.getcwd()
        context_filename = os.path.join(repo_root, "gke_mcp", "install", "GEMINI.md")
        if not os.path.isfile(context_filename):
            context_filename = os.path.join(repo_root, "pkg", "install", "GEMINI.md")
        if not os.path.isfile(context_filename):
            raise FileNotFoundError(f"Could not read context file at {context_filename}")

    extension_dir = os.path.join(opts.install_dir, ".gemini", "extensions", "gke-mcp")
    os.makedirs(extension_dir, mode=0o750, exist_ok=True)

    manifest = {
        "name": "gke-mcp",
        "version": opts.version,
        "description": "Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.",
        "contextFileName": context_filename,
        "mcpServers": {
            "gke": {
                "command": opts.exe_path
            }
        }
    }

    manifest_path = os.path.join(extension_dir, "gemini-extension.json")
    with open(manifest_path, "w", encoding="utf-8") as f:
        json.dump(manifest, f, indent=2)

    if not opts.developer_mode:
        gemini_md_path = os.path.join(extension_dir, "GEMINI.md")
        with open(gemini_md_path, "w", encoding="utf-8") as f:
            f.write(get_gemini_markdown())

def install_cursor_mcp_extension(opts: Options) -> None:
    mcp_dir = os.path.join(opts.install_dir, ".cursor")
    os.makedirs(mcp_dir, mode=0o750, exist_ok=True)

    mcp_path = os.path.join(mcp_dir, "mcp.json")
    config: Dict[str, Any] = {}

    if os.path.isfile(mcp_path):
        try:
            with open(mcp_path, "r", encoding="utf-8") as f:
                config = json.load(f)
        except Exception as e:
            raise RuntimeError(f"Could not parse existing MCP configuration: {e}")

    if "mcpServers" not in config:
        config["mcpServers"] = {}

    mcp_servers = config["mcpServers"]
    if not isinstance(mcp_servers, dict):
        logger.warning("mcpServers in Cursor MCP config is not a dictionary. Overwriting.")
        config["mcpServers"] = {}
        mcp_servers = config["mcpServers"]

    mcp_servers["gke-mcp"] = {
        "command": opts.exe_path,
        "type": "stdio"
    }

    with open(mcp_path, "w", encoding="utf-8") as f:
        json.dump(config, f, indent=2)

    rules_dir = os.path.join(mcp_dir, "rules")
    os.makedirs(rules_dir, mode=0o750, exist_ok=True)

    rule_content = CURSOR_RULE_HEADER + get_gemini_markdown()
    rule_path = os.path.join(rules_dir, "gke-mcp.mdc")
    with open(rule_path, "w", encoding="utf-8") as f:
        f.write(rule_content)

def install_claude_desktop_extension(opts: Options) -> None:
    config_path = get_claude_desktop_config_path()
    os.makedirs(os.path.dirname(config_path), mode=0o750, exist_ok=True)

    config: Dict[str, Any] = {}
    if os.path.isfile(config_path):
        try:
            with open(config_path, "r", encoding="utf-8") as f:
                config = json.load(f)
        except Exception as e:
            raise RuntimeError(f"Could not parse existing Claude Desktop config: {e}")

    if "mcpServers" not in config:
        config["mcpServers"] = {}

    mcp_servers = config["mcpServers"]
    if not isinstance(mcp_servers, dict):
        config["mcpServers"] = {}
        mcp_servers = config["mcpServers"]

    mcp_servers["gke-mcp"] = {
        "command": opts.exe_path
    }

    with open(config_path, "w", encoding="utf-8") as f:
        json.dump(config, f, indent=2)

def get_claude_desktop_config_path() -> str:
    if sys.platform == "darwin":
        config_dir = os.path.expanduser("~/Library/Application Support/Claude")
    elif sys.platform == "win32":
        app_data = os.getenv("APPDATA", "")
        if not app_data:
            raise RuntimeError("APPDATA environment variable not set")
        config_dir = os.path.join(app_data, "Claude")
    else:  # Assume linux
        config_dir = os.path.expanduser("~/.config/Claude")

    return os.path.join(config_dir, "claude_desktop_config.json")

def install_claude_code_extension(opts: Options) -> None:
    claude_md_path = os.path.join(opts.install_dir, "CLAUDE.md")
    exists = os.path.isfile(claude_md_path)

    if exists:
        print("Warning: CLAUDE.md already exists. The GKE MCP usage instructions will be appended.")
    else:
        print("Note: CLAUDE.md does not exist. A new one will be created.")

    response = input("Would you like to proceed? (yes/no): ")
    if response.strip().lower() != "yes":
        print("Installation canceled.")
        return

    usage_guide_path = os.path.join(opts.install_dir, "GKE_MCP_USAGE_GUIDE.md")
    with open(usage_guide_path, "w", encoding="utf-8") as f:
        f.write(get_gemini_markdown())
    print("Created GKE_MCP_USAGE_GUIDE.md.")

    claude_line = f"\n# GKE-MCP Server Instructions\n - @{usage_guide_path}\n"
    with open(claude_md_path, "a", encoding="utf-8") as f:
        f.write(claude_line)
    print("Added a reference to GKE_MCP_USAGE_GUIDE.md in CLAUDE.md.")

    # Execute the command to add the MCP server
    cmd = ["claude", "mcp", "add", "gke-mcp", opts.exe_path]
    print(f"Executing command: {' '.join(cmd)}")
    try:
        subprocess.run(cmd, check=True)
    except Exception as e:
        raise RuntimeError(f"Failed to run command 'claude mcp add': {e}")
