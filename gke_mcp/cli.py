import click
import logging
import sys
from gke_mcp.config import Config
from gke_mcp.install import (
    Options,
    install_gemini_cli_extension,
    install_cursor_mcp_extension,
    install_claude_desktop_extension,
    install_claude_code_extension
)

VERSION = "0.1.0"

# Setup basic logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    stream=sys.stderr
)
logger = logging.getLogger("gke-mcp.cli")

@click.group(invoke_without_command=True)
@click.option("--server-mode", default="stdio", type=click.Choice(["stdio", "http"]), help="transport to use for the server: stdio or http")
@click.option("--server-host", default="127.0.0.1", help="server host to use when server-mode is http")
@click.option("--server-port", default=8080, type=int, help="server port to use when server-mode is http")
@click.option("--allowed-origins", default="http://localhost", help="comma-separated list of allowed Origin headers")
@click.option("--enable-delete-tools", is_flag=True, default=False, help="Enable destructive delete tools (delete_cluster, delete_node_pool)")
@click.version_option(version=VERSION)
@click.pass_context
def main(ctx, server_mode, server_host, server_port, allowed_origins, enable_delete_tools):
    """An MCP Server for Google Kubernetes Engine (GKE)"""
    if ctx.invoked_subcommand is None:
        # Default behavior: start server
        origins = [o.strip() for o in allowed_origins.split(",") if o.strip()]
        logger.info(f"Starting GKE MCP Server ({VERSION}) in mode '{server_mode}'")
        
        # Construct config
        cfg = Config(version=VERSION, enable_delete_tools=enable_delete_tools)
        
        try:
            from gke_mcp.server import start_server
            start_server(cfg, server_mode, server_host, server_port, origins)
        except Exception as e:
            logger.exception("Server error occurred")
            sys.exit(1)

@main.group(name="install")
def install():
    """Install the GKE MCP Server into your AI tool settings."""
    pass

@install.command(name="gemini-cli")
@click.option("-d", "--developer", is_flag=True, default=False, help="Install the MCP Server in developer mode for Gemini CLI")
@click.option("-p", "--project-only", is_flag=True, default=False, help="Install the MCP Server only for the current project")
def install_gemini(developer, project_only):
    """Install the GKE MCP Server into your Gemini CLI settings."""
    try:
        opts = Options(version=VERSION, project_only=project_only, developer_mode=developer)
        install_gemini_cli_extension(opts)
        click.echo("Successfully installed GKE MCP server as a gemini-cli extension.")
    except Exception as e:
        logger.error(f"Failed to install for gemini-cli: {e}")
        sys.exit(1)

@install.command(name="cursor")
@click.option("-p", "--project-only", is_flag=True, default=False, help="Install the MCP Server only for the current project")
def install_cursor(project_only):
    """Install the GKE MCP Server into your Cursor settings."""
    try:
        opts = Options(version=VERSION, project_only=project_only, developer_mode=False)
        install_cursor_mcp_extension(opts)
        click.echo("Successfully installed GKE MCP server as a cursor MCP server.")
    except Exception as e:
        logger.error(f"Failed to install for cursor: {e}")
        sys.exit(1)

@install.command(name="claude-desktop")
def install_claude_desktop():
    """Install the GKE MCP Server into your Claude Desktop settings."""
    try:
        opts = Options(version=VERSION, project_only=False, developer_mode=False)
        install_claude_desktop_extension(opts)
        click.echo("Successfully installed GKE MCP server in Claude Desktop configuration.")
    except Exception as e:
        logger.error(f"Failed to install for Claude Desktop: {e}")
        sys.exit(1)

@install.command(name="claude-code")
@click.option("-p", "--project-only", is_flag=True, default=False, help="Install the MCP Server only for the current project")
def install_claude_code(project_only):
    """Install the GKE MCP Server into your Claude Code CLI settings."""
    try:
        opts = Options(version=VERSION, project_only=project_only, developer_mode=False)
        install_claude_code_extension(opts)
        click.echo("Successfully installed GKE MCP server for Claude Code.")
    except Exception as e:
        logger.error(f"Failed to install for Claude Code: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
