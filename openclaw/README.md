# GKE MCP Server - OpenClaw Integration

This directory contains the integration components for bringing the power of the Google Kubernetes Engine (GKE) Model Context Protocol (MCP) server directly into the [OpenClaw](https://openclaw.ai/) ecosystem.

## What is Installed?

> **🚀 UPCOMING FEATURE: Live MCP Server Integration**
> *While the GKE subagents and skills install successfully today, direct integration with the `gke-mcp` MCP Server binary is an upcoming feature. Full real-time cluster interaction capabilities will be enabled in a future update. In the meantime, agents can still perform cluster operations using standard `gcloud` and `kubectl` shell commands.*

When you run the installation script, it enriches your OpenClaw environment with a suite of GKE management capabilities:

1. **GKE MCP Server (`gke-mcp`)**
   - The core binary is installed to your local system and automatically registered as an MCP server within OpenClaw.
   - It exposes a rich set of tools to read cluster states, inspect resources, view logs, deploy manifests, and run diagnostics directly against your GKE environment.

2. **Specialized Subagents (e.g., `operator`)**
   - The installer creates dedicated, isolated AI subagents tailored for specific GKE workflows.
   - The primary agent, **GKE Operator** (`operator`), comes pre-configured with a custom identity (`IDENTITY.md`) and operational persona (`SOUL.md`), ensuring it acts as a knowledgeable, safety-conscious cluster operator.

3. **Domain-Specific Skills**
   - Agents are provisioned with targeted "Skills" (expert instructions and workflows) downloaded dynamically into their workspace. 
   - For example, the `operator` agent comes equipped with the `gke-observability` skill out-of-the-box, providing it with structured playbooks for monitoring, metric analysis, and log troubleshooting.

4. **Semantic Routing Configuration**
   - The integration automatically patches OpenClaw's configuration to allow seamless semantic routing.
   - This means OpenClaw's main gateway can automatically detect GKE-related queries and route them directly to the `operator` expert agent without manual intervention.

## Installation

You can install and configure the entire integration (binary, agents, skills, and configuration) using a single command:

```bash
curl -fsSL https://raw.githubusercontent.com/GoogleCloudPlatform/gke-mcp/main/openclaw/scripts/install.sh | bash
```

> **Note for Contributors:** If you are testing from a fork or custom branch, replace `GoogleCloudPlatform` and `main` with your GitHub username and branch name.

## Getting Started

Once installation is complete, restart your OpenClaw gateway if it is already running. You can interact with your new GKE expert immediately through your standard OpenClaw TUI or configured channels.

To start a session directly with a subagent (for example, the GKE Operator agent), use:

```bash
openclaw tui --session agent:operator:main
```

## References

- [OpenClaw Documentation](https://docs.openclaw.ai/)
- [Building OpenClaw Plugins](https://docs.openclaw.ai/plugins/building-plugins)
