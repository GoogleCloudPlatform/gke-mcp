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

- `cluster_toolkit`: Creates AI optimized GKE Clusters.
- `list_clusters`: List your GKE clusters.
- `get_cluster`: Get detailed about a single GKE Cluster.
- `giq_generate_manifest`: Generate a GKE manifest for AI/ML inference workloads using Google Inference Quickstart.
- `list_recommendations`: List recommendations for your GKE clusters.
- `list_mr_schema`: List all Monitoring Resources(including logging, metrics etc) for your project.

## Context 

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
    ./gke-mcp install gemini-cli
    ```

    This will make `gemini-cli` use your locally compiled binary.

