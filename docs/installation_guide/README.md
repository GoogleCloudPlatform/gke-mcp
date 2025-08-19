# Installation Guides for GKE MCP

This directory contains detailed instructions on how to install and configure the GKE MCP Server with different AI clients.

- **[Gemini CLI](../../README.md#add-the-mcp-server-to-your-ai)**
- **[Cursor](install_cursor.md)**
- **[Claude Applications](install_claude.md)**

## Other AIs

For AIs that support JSON configuration, usually you can add the MCP server to your existing config with the below JSON. Don't copy and paste it as-is, merge it into your existing JSON settings.

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp"
    }
  }
}
```
