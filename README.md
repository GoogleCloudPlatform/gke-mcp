# GKE MCP Server

Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.

# Setup

Clone this repo and add the following to your AI tool. For [Gemini CLI](https://github.com/google-gemini/gemini-cli) the file is ~/.gemini/settings.json.

```json
"mcpServers":{
  "gke": {
    "cwd": "<CLONE DIR>/gke-mcp",
    "command": "go",
    "args": ["run", "server.go"]
  }
}
```

## Tools

- `list_clusters`: List your GKE clusters.
- `get_cluster`: Get detailed about a single GKE Cluster.
