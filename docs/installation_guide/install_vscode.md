# Installing the GKE MCP Server in Visual Studio Code

This guide provides detailed steps on how to install and configure the GKE MCP Server for use with Visual Studio Code. This allows you to leverage Visual Studio Code's AI capabilities to interact with your GKE clusters using natural language prompts.

## Prerequisites

The GKE MCP Server is a command-line tool. You must have the binary installed on your system before configuring it in Visual Studio Code.

Please follow the [installation instructions in the main readme](../../README.md#install-the-mcp-server) to install the `gke-mcp` binary.

## Installing via Visual Studio Code MCP Extension

The recommended way to use MCP servers in Visual Studio Code is through the official MCP Client extension or similar MCP-enabled extensions.

### Step 1: Install the MCP Extension

1. Open Visual Studio Code
2. Go to Extensions (Ctrl+Shift+X or Cmd+Shift+X)
3. Search for "MCP Client" or "Model Context Protocol"
4. Install the official MCP Client extension

### Step 2: Configure the MCP Server

After installing the MCP extension, you need to configure it to use the GKE MCP Server.

#### Method 1: Extension Settings

1. Open Visual Studio Code Settings (Ctrl+, or Cmd+,)
2. Search for "MCP" or "Model Context Protocol"
3. Find the MCP Servers configuration section
4. Add a new server configuration:

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp",
      "type": "stdio"
    }
  }
}
```

#### Method 2: JSON Configuration File

Some MCP extensions support a JSON configuration file. Create or edit the MCP configuration file:

- **Global**: `~/.vscode/mcp.json`
- **Workspace**: `.vscode/mcp.json` in your project root

Add the following configuration:

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp",
      "type": "stdio"
    }
  }
}
```

### Step 3: Verify the Connection

1. Restart Visual Studio Code after configuring the MCP server
2. Open the MCP extension panel (usually in the sidebar)
3. Look for "gke-mcp" in the list of configured servers
4. A green indicator or "Connected" status confirms successful setup

## Using GKE MCP in Visual Studio Code Chat

Once connected, you can use the Visual Studio Code chat interface to interact with your GKE environment:

- **Prompt:** "List all my GKE clusters in us-central1"
- **Expected Behavior:** The AI assistant will use the `list_clusters` tool to retrieve and display cluster information

### Example Prompts

- "What is the current status of my cluster named 'production-cluster'?"
- "Get the kubeconfig for my development cluster"
- "Show me the node pools in my GKE cluster"
- "Are there any recommendations available for my clusters?"

## Configuration Options

### Command Path

If the `gke-mcp` command is not in your system's PATH, provide the full path:

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "/full/path/to/gke-mcp",
      "type": "stdio"
    }
  }
}
```

### Environment Variables

Some MCP extensions support environment variables. You can pass additional environment configuration if needed:

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp",
      "type": "stdio",
      "env": {
        "CLOUDSDK_CORE_PROJECT": "your-project-id"
      }
    }
  }
}
```

## Troubleshooting

### Server Fails to Start

- Verify the `gke-mcp` binary is installed and executable
- Check that the command path is correct in the configuration
- Ensure GCP credentials are properly configured (`gcloud auth application-default login`)

### Connection Timeout

- Check your network connectivity to Google Cloud APIs
- Verify the GCP project and region settings
- Review Visual Studio Code extension logs for detailed error messages

### Extension Not Showing MCP Servers

- Restart Visual Studio Code after configuration changes
- Check that the MCP extension is properly installed and enabled
- Verify the JSON configuration file is valid

### Logs and Debugging

- Check Visual Studio Code Developer Tools (Help > Toggle Developer Tools) for extension errors
- Look for MCP-related messages in the console output
- Verify GCP authentication with `gcloud auth list`

## Additional Resources

- [GKE MCP Server GitHub Repository](https://github.com/GoogleCloudPlatform/gke-mcp)
- [Visual Studio Code Extension Marketplace](https://marketplace.visualstudio.com/vscode)
- [Google Cloud SDK Documentation](https://cloud.google.com/sdk/docs)
