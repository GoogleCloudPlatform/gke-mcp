# Installing the GKE MCP Server in Cursor

This guide provides detailed steps on how to install and configure the GKE MCP Server for use with the Cursor IDE. This allows you to leverage Cursor's AI agent to interact with your GKE clusters using natural language prompts.

## Prerequisites and Installation of the `gke-mcp` Binary

The GKE MCP Server is a command-line tool. You must have the binary installed on your system before configuring it in Cursor.

Please follow the [installation instructions in the main readme](../../README.md#install-the-mcp-server) to install the `gke-mcp` binary.

## Configure `gke-mcp` as a Cursor MCP

Cursor uses a JSON configuration file to manage its MCP servers. You must define your server in this file.

- **For global use:** Edit the global configuration file at `~/.cursor/mcp.json`.
- **For project-specific use:** Create a `.cursor/mcp.json` file in your project's root directory.

Add the following configuration snippet to your `mcp.json` file. If the file already exists, merge this into the `mcpServers` object.

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

## Adapting Context from `GEMINI.md`

A key challenge in this integration is that the `gke-mcp` tool relies on a `gemini.md` file for its system prompts. To avoid rewriting the core logic, we will adapt this file's functionality for Cursor by using the **"Rules"** system.

### Steps to Implement the Rule

1. **Create the Rule File**: Create a new file named `gke-mcp.mdc` in your project's `.cursor/rules/` directory.

2. **Add Metadata**: Add the following metadata block to the top of the `gke-mcp.mdc` file.

   ```markdown
   ---
   name: GKE MCP Instructions
   description: Provides guidance for using the gke-mcp tool.
   ---
   ```

3. **Copy Content**: Copy the entire content of the [`gke-mcp`'s `GEMINI.md`](../../pkg/install/GEMINI.md) file into your `gke-mcp.mdc` file, placing it directly below the metadata block.

This rule will be configured to be **Agent Requested** by default, allowing the AI to dynamically include the GKE context in its prompts only when it's relevant to your conversation.

## Verification and Usage

### How to Verify the Connection

1. Restart Cursor after modifying the configuration.
2. Open **Settings** (`Ctrl + ,` or `Cmd + ,`).
3. Navigate to **Features \> MCP**. A green dot next to the "gke-mcp" entry indicates a successful connection.

### Sample Usage in Cursor

Once connected, you can use natural language prompts in the Cursor chat to interact with your GKE environment. For example:

- **Prompt:** "List all the GKE clusters I have in the `us-central1` region."
- **Expected Behavior:** Cursor's AI will propose using the `list_clusters` tool. After your approval, it will execute the command and display the results in a readable format.
