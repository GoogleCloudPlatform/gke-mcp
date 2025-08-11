# Installation Guides for GKE MCP

This directory contains detailed instructions on how to install and configure the GKE MCP Server with different AI clients.

- **[Gemini CLI](../../README.md#installation)**
- **[Cursor](install_cursor.md)**
- **Claude Desktop**: instructions coming soon

#### Other AIs

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
