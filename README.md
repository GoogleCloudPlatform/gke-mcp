# GKE MCP Server

Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.

# Prerequisites

- [Go](https://go.dev/doc/install)
- A local MCP Client
  - [Gemini CLI](https://github.com/google-gemini/gemini-cli)

# Setup

Clone this repo and add the MCP server config.

## Gemini Setup

Run this command:
```sh
go run ./cmd --install_gemini
```

## Others

Add this MCP server config to your AI settings file:

```json
"mcpServers":{
  "gke": {
    "cwd": "<CLONE DIR>/gke-mcp",
    "command": "sh",
    "args": ["./run_mcp_server.sh"]
  }
}
```

# Tools

- `list_clusters`: List your GKE clusters.
- `get_cluster`: Get detailed about a single GKE Cluster.
- `giq_generate_manifest`: Generate a GKE manifest for AI/ML inference workloads using Google Inference Quickstart.
