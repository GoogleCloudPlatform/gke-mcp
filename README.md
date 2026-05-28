# GKE MCP Server and Gemini CLI Extension

Chat with your AI to manage Google Kubernetes Engine (GKE)

<img src="https://raw.githubusercontent.com/GoogleCloudPlatform/gke-mcp/main/assets/gke-mcp-gemini-cli-demo.gif" alt="A demonstration of using the GKE MCP server with the Gemini CLI" width="600">

---

## Overview

GKE MCP (Model Context Protocol) is **Google's agentic way** to manage GKE. It connects **Gemini, Claude, Codex, and more** to GKE, allowing you to manage clusters and troubleshoot issues using plain words.

This project is a comprehensive bundle including:
- **MCP Tools & Resources:** Standardized actions for GKE management.
- **Expert Skills:** Specialized logic that is more reliable than pre-trained LLMs and includes Google's GKE best practices.
- **Plugin/Extension:** A ready-to-use bundle for Gemini CLI that combines skills, resources, and tools.

---

## Why this, over out-of-the-box AI?

Standard AI agents can run terminal commands, capable of many straight forward user instructions. Meanwhile GKE MCP is the **Google-recommended way** to manage GKE, featuring:

- **Expert Logic:** Built-in intelligence for GKE cost allocation and upgrade paths.
- **Reliability:** Specialized Skills outperforming generic LLM reasoning.
- **Safety:** Uses validated tools instead of raw shell commands to reduce misconfiguration.
- **Up-to-Date:** Supports the latest GKE features and security standards.
- **Context:** Provides AI with known issues and release notes for accurate troubleshooting.

---

## How to use it

### 1. Connect to your AI Agent
Detailed instructions for popular clients:
- [Gemini CLI](docs/installation_guide/gemini-cli.md)
- [Cursor](docs/installation_guide/cursor.md)
- [Claude (Desktop & Code)](docs/installation_guide/claude.md)
- [Visual Studio Code](docs/installation_guide/vscode.md)

### 2. As a Gemini CLI Extension
This installs the full bundle (Skills + Resources + MCP Tools).

1.  [Install Gemini CLI](https://github.com/google-gemini/gemini-cli?tab=readme-ov-file#-installation).
2.  Install extension:
    ```sh
    gemini extensions install https://github.com/GoogleCloudPlatform/gke-mcp.git
    ```

### 3. As an MCP Server
Integrate GKE into **Claude, Cursor, VS Code,** or custom AI apps via the [MCP protocol](https://modelcontextprotocol.io/).

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

### 4. As Standalone Skills
Install specific skills without the full toolset. These are more reliable than standard LLM responses as they contain codified expert logic.

There are several ways to install these skills:

1. **Automatic Detection**: When you install the MCP server as a
   [Gemini CLI Extension](#2-as-a-gemini-cli-extension), the CLI automatically
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

---

## What's inside?

### MCP Tools
Actions your AI can perform on your GKE infrastructure:
- `cluster_toolkit_download`: Download the Cluster Toolkit Git repository.
- `list_clusters`: List GKE clusters.
- `get_cluster`: Get detailed information about a single GKE cluster.
- `create_cluster`: Create a new GKE cluster (defaults to Autopilot).
- `get_kubeconfig`: Configure kubeconfig for a GKE cluster.
- `update_cluster`: Update a GKE cluster.
- `get_node_sos_report`: Generate and download an SOS report from a GKE node.
- `delete_cluster`: Delete a GKE cluster (if enabled).
- `list_node_pools`: List node pools in a GKE cluster.
- `get_node_pool`: Get details for a GKE node pool.
- `create_node_pool`: Create a new node pool in a GKE cluster.
- `update_node_pool`: Update a GKE node pool.
- `delete_node_pool`: Delete a GKE node pool (if enabled).
- `gke_deploy`: Deploy a workload to a GKE cluster using a configuration file.
- `query_logs`: Query Google Cloud Platform logs using Logging Query Language (LQL).
- `get_log_schema`: Get the schema for a specific GKE log type.
- `list_monitored_resource_descriptors`: List monitored resource descriptors for GKE.
- `list_recommendations`: List recommendations for GKE clusters.
- `get_k8s_changelog`: Get Kubernetes changelog for upgrades.
- `get_gke_release_notes`: Get GKE release notes.
- `generate_manifest`: Generate a Kubernetes manifest using Vertex AI.
- `get_k8s_resource`: Gets one or more Kubernetes resources from a cluster.
- `list_k8s_events`: Retrieves events from a Kubernetes cluster.
- `get_k8s_version`: Retrieves the Kubernetes server version for a given cluster.
- `apply_k8s_manifest`: Applies a Kubernetes manifest to a cluster using server-side apply.
- `get_k8s_logs`: Gets logs from a Kubernetes container in a pod.
- `delete_k8s_resource`: Delete a Kubernetes resource from a cluster.

### MCP Prompts
Guided templates for complex workflows:
- `gke:cost`: Answer natural language questions about GKE-related costs.
- `gke:deploy`: Deploys a workload to a GKE cluster using a configuration file.
- `gke:upgrade-risk-report`: GKE control plane upgrade risk report.
- `gke:upgrades-best-practices-risk-report`: GKE control plane upgrade best practices.

### MCP Context
Expert data provided to the AI:
- **Cost**: Interpretation of billing data for clusters, namespaces, and workloads.
- **GKE Known Issues**: Automatic checking against documented bugs.

### Available Skills
Specialized workflows that are more reliable and safer than generic LLM reasoning:
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

---

## Supported MCP Transports

By default, `gke-mcp` uses the [stdio](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports#stdio) transport. Additionally, the [Streamable HTTP](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports#streamable-http) transport is supported as well.

```sh
# Run as an HTTP server
gke-mcp --server-mode http --server-port 8080
```

### Advanced Configuration
To connect Gemini CLI to a remote HTTP server, update your `~/.gemini/settings.json`:
```json
{
  "mcpServers": {
    "gke": {
      "httpUrl": "http://127.0.0.1:8080/mcp"
    }
  }
}
```

---

## Development

To compile the binary and update the `gemini-cli` extension with your local changes, follow these steps:

1. Remove the global gke-mcp configuration
   ```sh
   rm -rf ~/.gemini/extensions/gke-mcp
   ```
2. Build the binary from the root of the project:
   ```sh
   go build -o gke-mcp .
   ```
3. Run the installation command to update the extension manifest:
   ```sh
   ./gke-mcp install gemini-cli --developer
   ```

See [Contributing Guide](contributing.md).

---

## Disclaimers

- The Google Cloud Platform Terms of Service (available at [https://cloud.google.com/terms/](https://cloud.google.com/terms/)) and the Data Processing and Security Terms (available at [https://cloud.google.com/terms/data-processing-terms](https://cloud.google.com/terms/data-processing-terms)) do not apply to any component of the GKE MCP Server software.
- This tool is provided for education and experimentation, and is not an officially supported Google product. It is maintained on a best-effort basis, and may change without notice.
- This project interacts with Large Language Models and comes with inherent risks.
  - **Use at Your Own Risk:** This software is experimental, non-deterministic, and provided "AS IS" with NO GUARANTEES or warranties.
  - **NOT FOR PRODUCTION USE.**
  - **Data Sensitivity:** Avoid using untrusted data. NEVER input secrets, API keys, or sensitive information.
  - **Verify Outputs:** LLM responses can be unpredictable and may be inaccurate. Always verify results.
