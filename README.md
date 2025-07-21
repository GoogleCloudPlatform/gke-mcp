# GKE MCP Server

Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.

# Installation

Choose a way to install the MCP Server and then connect your AI to it.

## Install the MCP Server

#### Quick Install (Linux & MacOS only)

```sh
curl -sSL https://raw.githubusercontent.com/GoogleCloudPlatform/gke-mcp/main/install.sh | bash
```

#### Manual Install

```sh
go install github.com/GoogleCloudPlatform/gke-mcp@latest
```

The `gke-mcp` binary will be installed in the directory specified by the `GOBIN` environment variable. If `GOBIN` is not set, it defaults to `$GOPATH/bin` and, if `GOPATH` is also not set, it falls back to `$HOME/go/bin`.

You can find the exact location by running `go env GOBIN`. If the command returns an empty value, run `go env GOPATH` to find the installation directory.

## Add the MCP Server to your AI

#### Gemini CLI

Install it as a `gemini-cli` extension:

```sh
gke-mcp install gemini-cli
```

This will create a manifest file in `./.gemini/extensions/gke-mcp` that points to the `gke-mcp` binary.

#### Other AIs

For AIs that support JSON configuration, usually you can add the MCP server to your existing config with the below JSON. Don't copy and paste it as-is, merge it into your existing JSON settings.

```json
{
  "mcpServers": {
    "gke-mcp": {
      "command": "gke-mcp",
    }
  }
}
```

## MCP Tools

- `cluster_toolkit`: Creates AI optimized GKE Clusters.
- `list_clusters`: List your GKE clusters.
- `get_cluster`: Get detailed about a single GKE Cluster.
- `giq_generate_manifest`: Generate a GKE manifest for AI/ML inference workloads using Google Inference Quickstart.
- `list_recommendations`: List recommendations for your GKE clusters.

## MCP Context 

In addition to the tools above, a lot of value is provided through the bundled context instructions.

- **Cost**: The provided instructions allows the AI to answer many questions related to GKE costs, including queries related to clusters, namespaces, and Kubernetes workloads.

## Development

To compile the binary and update the `gemini-cli` extension with your local changes, follow these steps:

1.  Build the binary from the root of the project:

    ```sh
    go build -o gke-mcp .
    ```

2.  Run the installation command to update the extension manifest:

    ```sh
    ./gke-mcp install gemini-cli --developer
    ```

    This will make `gemini-cli` use your locally compiled binary.
