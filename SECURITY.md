# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the most recent version of the GKE MCP Server. We encourage users to keep their installations up to date with the latest releases.

| Version  | Supported          |
| -------- | ------------------ |
| Latest   | :white_check_mark: |
| < Latest | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Google takes the security of our software products and services seriously, including all source code repositories managed through our GitHub organizations.

If you believe you have found a security vulnerability in the GKE MCP Server, please report it to us through coordinated disclosure.

### Reporting Process

**Please include the following information in your report:**

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Where to Report

Please report security vulnerabilities by emailing:

**[security@google.com](mailto:security@google.com)**

Alternatively, you can use Google's Vulnerability Reward Program (VRP):

**<https://bughunters.google.com/>**

### What to Expect

- You should receive an acknowledgment of your report within 24-48 hours.
- We will investigate the issue and determine its severity and impact.
- We will keep you informed of our progress.
- Once the issue is resolved, we will publicly disclose the vulnerability (with credit to you, if desired).

## Security Best Practices

When using the GKE MCP Server, we recommend following these security best practices:

### Authentication

- **Use Application Default Credentials (ADC)**: The GKE MCP Server uses Google Cloud's Application Default Credentials. Always authenticate using `gcloud auth application-default login`.
- **Never commit credentials**: Do not commit API keys, service account keys, or other credentials to version control.
- **Use service accounts**: For production use, use service accounts with minimal required permissions.

### Network Security

- **HTTP Mode Warning**: When using HTTP transport mode (`--server-mode http`), the server listens on all network interfaces. Use firewalls and network security controls to restrict access.
- **CORS Configuration**: Configure `--allowed-origins` appropriately to restrict which origins can access your MCP server.
- **Use stdio mode for local use**: The default stdio transport is more secure for local AI tool integration.

### Data Handling

- **Sensitive Data**: Avoid sending sensitive or confidential information to Large Language Models through the MCP server.
- **Secrets Management**: Never input API keys, passwords, or other secrets into AI prompts.
- **Verify AI Outputs**: Always verify LLM-generated configurations and commands before applying them to production systems.

### Access Control

- **Principle of Least Privilege**: Grant the minimum GCP IAM permissions necessary for your use case.
- **Audit Logging**: Enable and monitor GCP audit logs for API calls made through the MCP server.
- **Project Isolation**: Use separate GCP projects for development, testing, and production environments.

## Security Updates

We will announce security updates through:

- GitHub Security Advisories on this repository
- Release notes in the [Releases](https://github.com/GoogleCloudPlatform/gke-mcp/releases) section
- Commit messages tagged with `[SECURITY]`

## Disclaimer

This tool is provided for education and experimentation and is not an officially supported Google product. It is maintained on a best-effort basis. For official Google Cloud support, please refer to [Google Cloud Support](https://cloud.google.com/support).

## Additional Resources

- [Google Cloud Security Best Practices](https://cloud.google.com/security/best-practices)
- [GKE Security Best Practices](https://cloud.google.com/kubernetes-engine/docs/how-to/hardening-your-cluster)
- [Google's Vulnerability Reward Program](https://bughunters.google.com/)
- [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials)
