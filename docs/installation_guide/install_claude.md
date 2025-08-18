# Installing the GKE MCP Server in Claude Applications

This guide covers installation of the GKE MCP server for Claude Desktop, Claude Code CLI, and Claude Web applications.

## Claude Desktop

Claude Desktop provides a graphical interface for interacting with the GKE MCP Server.

### Prerequisites

1. Confirm the `gke-mcp` binary is installed. If not, please follow the [installation instructions in the main readme](../../README.md#install-the-mcp-server)
2. Claude Desktop is installed. If not, the application can be downloaded from [Claude's official site](https://claude.ai/download).

### Automatic Installation

The easiest way to install the GKE MCP Server for Claude Desktop is using the built-in installation command:
```commandline
gke-mcp install claude-desktop
```
After running the command, restart Claude Desktop for the changes to take effect.

### Manual Installation

If you prefer to configure Claude Desktop manually or the automatic installation failed, you can edit the 
configuration file directly.

#### Configuration File Location

Claude Desktop requires you to manually edit its configuration file, `claude_desktop_config.json`.
The location of the file varies by operating system:

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json` (unofficial support)

You can also find this file by going to the settings in the Claude Desktop app and looking for the Developer tab. There should be a button to edit config.

#### Installation

Open `claude_desktop_config.json` in a text editor. Then, find the mcpServers section within the JSON file. If it doesn't exist,
create it. Add the following JSON snippet, making sure to merge it correctly with any existing configurations. The command field
should point to the full path of the `gke-mcp` binary.

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp"
    }
  }
}
```

Note: If the `gke-mcp` command is not in your system's PATH, you must provide the full path to the binary.

#### Troubleshooting

- Check logs at:
  - **macOS**: `~/Library/Logs/Claude/`
  - **Windows**: `%APPDATA%\Claude\logs\`
- Look for `mcp-server-gke-mcp.log` for server-specific errors
- Ensure configuration file is valid JSON

## Claude Code CLI

Claude Code CLI provides command-line access to Claude with MCP server integration.

Installation steps coming soon.

## Claude Web (claude.ai)

Claude Web supports remote MCP servers through the Integrations built-in feature.

Installation steps coming soon.
