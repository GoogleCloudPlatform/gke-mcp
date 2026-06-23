# Installation Guides for GKE MCP

This directory contains detailed instructions on how to install and configure the GKE MCP Server with different AI clients.

- **[Gemini CLI](../../README.md#add-the-mcp-server-to-your-ai)**
- **[Cursor](install_cursor.md)**
- **[Claude Applications](install_claude.md)**
- **[Visual Studio Code](install_vscode.md)**

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

## Configuring the Developer Knowledge API

The manifest generation agent (`generate_manifest` tool) can retrieve official GKE documentation, required annotations, and best practices using the Developer Knowledge API.

To enable this capability, you need to:

1. **Enable the API:** Enable the **Developer Knowledge API** in your Google Cloud Project (refer to the [Developer Knowledge API setup guide](https://developers.google.com/knowledge/api#enable_the_api) for details).
2. **Generate an API Key:** Create an API key with permissions to call the Developer Knowledge API.
3. **Configure the Environment Variable:** Set the `DK_API_KEY` environment variable in your client's MCP configuration.

For example, in Cursor or Claude Desktop, add the `env` field to the server configuration:

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp",
      "env": {
        "DK_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### Configuration Options

- `DK_API_KEY` (Required): The API key used to authenticate with the Developer Knowledge API.
- `DK_BASE_URL` (Optional): The base URL of the Developer Knowledge API (defaults to `https://knowledge.googleapis.com`).
