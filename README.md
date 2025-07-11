# GKE MCP Server

Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.

# Installation

1.  Install the tool:

    ```sh
    go install github.com/GoogleCloudPlatform/gke-mcp@latest
    ```

    The `gke-mcp` binary will be installed in the directory specified by the `GOBIN` environment variable. If `GOBIN` is not set, it defaults to `$GOPATH/bin` and, if `GOPATH` is also not set, it falls back to `$HOME/go/bin`.

    You can find the exact location by running `go env GOBIN`. If the command returns an empty value, run `go env GOPATH` to find the installation directory.

2.  Install it as a `gemini-cli` extension:

    ```sh
    gke-mcp install gemini-cli
    ```

    This will create a manifest file in `./.gemini/extensions/gke-mcp` that points to the installed `gke-mcp` binary.

## Tools

- `list_clusters`: List your GKE clusters.
- `get_cluster`: Get detailed about a single GKE Cluster.
- `giq_generate_manifest`: Generate a GKE manifest for AI/ML inference workloads using Google Inference Quickstart.
