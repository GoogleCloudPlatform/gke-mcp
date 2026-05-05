# GKE MCP Server and Gemini CLI Extension

Enable MCP-compatible AI agents to interact with Google Kubernetes Engine.

<img src="https://raw.githubusercontent.com/GoogleCloudPlatform/gke-mcp/main/assets/gke-mcp-gemini-cli-demo.gif" alt="A demonstration of using the GKE MCP server with the Gemini CLI" width="600">

## Installation

Choose a way to install the MCP Server and then connect your AI to it.

### Use as a Gemini CLI Extension

1. Install [Gemini CLI](https://github.com/google-gemini/gemini-cli?tab=readme-ov-file#-installation).

2. Install the extension

```sh
gemini extensions install https://github.com/GoogleCloudPlatform/gke-mcp.git
```

### Use in MCP Clients / Other AIs

#### Quick Install (Linux & macOS only)

```sh
curl -sSL https://raw.githubusercontent.com/GoogleCloudPlatform/gke-mcp/main/install.sh | bash
```

#### Manual Install

If you haven't already installed Go, follow [these instructions](https://go.dev/doc/install).

Once Go is installed, run the following command to install gke-mcp:

```sh
go install github.com/GoogleCloudPlatform/gke-mcp@latest
```

The `gke-mcp` binary will be installed in the directory specified by the `GOBIN` environment variable. If `GOBIN` is not set, it defaults to `$GOPATH/bin` and, if `GOPATH` is also not set, it falls back to `$HOME/go/bin`.

You can find the exact location by running `go env GOBIN`. If the command returns an empty value, run `go env GOPATH` to find the installation directory.

For additional help, refer to the troubleshoot section: [gke-mcp: command not found](TROUBLESHOOTING.md#gke-mcp-command-not-found-on-macos-or-linux).

### Add the MCP Server to your AI

For detailed instructions on how to connect the GKE MCP Server to various AI clients, including cursor, Visual Studio Code, and claude desktop, please refer to our dedicated [installation guide](docs/installation_guide/).

## MCP Tools

- `cluster_toolkit_download`: Download the Cluster Toolkit Git repository.
- `list_clusters`: List GKE clusters.
- `get_cluster`: Get detailed information about a single GKE cluster.
- `create_cluster`: Create a new GKE cluster (defaults to Autopilot).
- `get_kubeconfig`: Configure kubeconfig for a GKE cluster.
- `update_cluster`: Update a GKE cluster.
- `get_node_sos_report`: Generate and download an SOS report from a GKE node.
- `delete_cluster`: Delete a GKE cluster (if enabled).
- `gke_deploy`: Deploy a workload to a GKE cluster using a configuration file.
- `query_logs`: Query Google Cloud Platform logs using Logging Query Language (LQL).
- `get_log_schema`: Get the schema for a specific GKE log type.
- `list_monitored_resource_descriptors`: List monitored resource descriptors for GKE.
- `list_recommendations`: List recommendations for GKE clusters.
- `get_k8s_changelog`: Get Kubernetes changelog for upgrades.
- `get_gke_release_notes`: Get GKE release notes.
- `generate_manifest`: Generate a Kubernetes manifest using Vertex AI.

## MCP Prompts

Prompts provide guided workflows and expert knowledge templates.

- `gke:cost`: Answer natural language questions about GKE-related costs.
- `gke:deploy`: Deploys a workload to a GKE cluster using a configuration file.
- `gke:upgrade-risk-report`: GKE control plane upgrade risk report, analyzing the potential risks of upgrading from its current version to the target version. Performs pre-upgrade checks, API deprecations scans, and more.
- `gke:upgrades-best-practices-risk-report`: GKE control plane upgrade best practices, applied for the specified cluster. Helps making upgrades uneventful.

## MCP Context

In addition to the tools above, a lot of value is provided through the bundled context instructions.

- **Cost**: The provided instructions allows the AI to answer many questions related to GKE costs, including queries related to clusters, namespaces, and Kubernetes workloads.

- **GKE Known Issues**: The provided instructions allows the AI to fetch the latest GKE Known issues and check whether the cluster is affected by one of these known issues.

## Supported MCP Transports

By default, `gke-mcp` uses the [stdio]("https://modelcontextprotocol.io/specification/2025-06-18/basic/transports#stdio") transport. Additionally, the [Streamable HTTP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports#streamable-http) transport is supported as well.

You can set the transport mode using the following options:

`--server-mode`: transport to use for the server: stdio (default) or http

`--server-port`: server port to use when server-mode is http or sse; defaults to 8080

```sh
gke-mcp --server-mode http --server-port 8080
```

> [!WARNING]
> When using the `Streamable HTTP` transport, the server listens on all network interfaces (e.g., `0.0.0.0`), which can expose it to any network your machine is connected to.
> Please ensure you have a firewall ad/or other security measures in place to restrict access if the server is not intended to be public.

### Connecting Gemini CLI to the HTTP Server

To connect Gemini CLI to the `gke-mcp` HTTP server, you need to configure the CLI to point to the correct endpoint. You can do this by updating your `~/.gemini/settings.json` file. For a basic setup without authentication, the file should look like this:

```json
{
  "mcpServers": {
    "gke": {
      "httpUrl": "http://127.0.0.1:8080/mcp"
    }
  }
}
```

This configuration tells Gemini CLI how to reach the gke-mcp server running on your local machine at port 8080.

## Skills

Skills provide specialized capabilities and workflows to your AI agent.

### Available Skills

- `custom-golden-image-discovery`: Discover golden base images for GKE custom nodes.
- `gke-ai-troubleshooting-skill-creation-guide`: Guide for building high-quality GKE troubleshooting skills.
- `gke-ai-troubleshooting-tpu-connection-failure-vbar-oom`: Diagnose and prevent TPU connection failures and OOMs.
- `gke-app-onboarding`: Workflows for containerizing and deploying applications to GKE.
- `gke-backup-dr`: Configure Backup for GKE and disaster recovery.
- `gke-cluster-creator`: Create GKE clusters using predefined templates.
- `gke-cluster-lifecycle`: Manage lifecycle and upgrades of GKE clusters.
- `gke-compute-class-creator`: Create GKE ComputeClass resources.
- `gke-cost-analysis`: Answer questions about GKE-related costs.
- `gke-cost-optimization`: Optimize costs for GKE clusters.
- `gke-inference-quickstart`: Deploy optimized AI/ML inference workloads on GKE.
- `gke-multi-tenancy`: Implement multi-tenancy and governance in GKE.
- `gke-networking-edge`: Configure edge networking, ingress, and security on GKE.
- `gke-observability`: Set up and audit observability on GKE.
- `gke-productionize`: Prepare applications and clusters for production.
- `gke-reliability`: Ensure high availability and reliability of GKE workloads.
- `gke-storage`: Manage storage in GKE clusters.
- `gke-workload-scaling`: Scale GKE workloads using HPA and VPA.
- `gke-workload-security`: Audit and harden the security of GKE workloads.

### Installing Skills

There are several ways to install these skills:

1. **Automatic Detection**: When you install the MCP server as a
   [Gemini CLI Extension](#use-as-a-gemini-cli-extension), the CLI automatically
   detects and enables all skills located in the `skills/` folder.

2. **Standalone Individual Skill**: Install a specific skill without the full
   MCP extension:

   ```sh
   gemini skills install https://github.com/GoogleCloudPlatform/gke-mcp --path skills/<skill-name>
   ```

   Replace `<skill-name>` with the name of a skill from the `skills/` directory
   (e.g., `gke-cost-analysis`).

3. **Standalone Bulk Link**: To enable all skills at once without installing
   the full MCP extension:
   ```sh
   git clone https://github.com/GoogleCloudPlatform/gke-mcp.git
   gemini skills link ./gke-mcp/skills
   ```

## Development

To compile the binary and update the `gemini-cli` extension with your local changes, follow these steps:

1. Remove the global gke-mcp configuration

   ```sh
   rm -rf ~/.gemini/extensions/gke-mcp
   ```

1. Build the binary from the root of the project:

   ```sh
   go build -o gke-mcp .
   ```

1. Run the installation command to update the extension manifest:

   ```sh
   ./gke-mcp install gemini-cli --developer
   ```

   This will make `gemini-cli` use your locally compiled binary.

## Disclaimers

- The Google Cloud Platform Terms of Service (available at [https://cloud.google.com/terms/](https://cloud.google.com/terms/)) and the Data Processing and Security Terms (available at [https://cloud.google.com/terms/data-processing-terms](https://cloud.google.com/terms/data-processing-terms)) do not apply to any component of the GKE MCP Server software.
- This tool is provided for education and experimentation, and is not an officially supported Google product. It is maintained on a best-effort basis, and may change without notice.
- This project interacts with Large Language Models and comes with inherent risks.
  - **Use at Your Own Risk:** This software is experimental, non-deterministic, and provided "AS IS" with NO GUARANTEES or warranties.
  - **NOT FOR PRODUCTION USE.**
  - **Data Sensitivity:** Avoid using untrusted data. NEVER input secrets, API keys, or sensitive information.
  - **Verify Outputs:** LLM responses can be unpredictable and may be inaccurate. Always verify results.
