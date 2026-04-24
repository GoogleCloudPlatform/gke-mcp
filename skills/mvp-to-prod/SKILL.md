# GKE MVP-to-Production Workflow

This skill outlines a production-grade workflow for promoting applications from an MVP state to a reliable, multi-environment production deployment on GKE.

## Core Principles

1.  **Environment Isolation**: Maintain strictly separate namespaces for `staging` and `prod`.
2.  **GitOps & PR-Based Workflow**: All changes (code or infrastructure) must go through a Pull Request. No direct commits to the `main` branch.
3.  **Standardized Builds**: Use Cloud Build for deterministic, reproducible container image builds.
4.  **Kustomize for Configuration**: Use Kustomize to manage environment-specific parameters (hostnames, resource limits, replicas) while keeping the base manifests dry.
5.  **Staging as a Quality Gate**: Applications must be stable in the `staging` namespace before promotion.

## Workflow Sequence

### 1. Development & Build
- Implement features/fixes in a feature branch.
- Create a Pull Request (PR) to the `main` branch.
- Once merged, trigger Cloud Build to create new versioned image tags.
- **Rule**: Avoid `latest` tags. Use semantic versioning or commit SHA-based tags for traceability.

### 2. Deployment to Staging
- Update the Kustomize overlay for the `staging` namespace with the new image tag.
- Apply the configuration: `kubectl apply -k k8s/overlays/staging`.
- **Validation**: Monitor the rollout for success. Check for `CrashLoopBackOff`, `ImagePullBackOff`, or probe failures.

### 3. Stability Monitoring
- Monitor the application in `staging` for a defined "Crucible" period (e.g., 6 hours).
- **Automation**: Use an automated agent or prober to verify health.
- **Rule**: Ignore transient node churn (e.g., Spot node preemptions) if the application recovers automatically. Only application-level crashes should reset the stability clock.

### 4. Promotion to Production
- Only after successful verification in `staging`, promote the exact same image and configuration logic to `prod`.
- Update the Kustomize overlay for the `prod` namespace: `kubectl apply -k k8s/overlays/prod`.
- **Window Management**: Execute production rollouts only within defined maintenance windows to minimize impact.

## Manifest Structure (Kustomize)

Maintain a clear hierarchy in your repository:
- `k8s/base/`: Contains generic resources (Deployment, Service, HPA).
- `k8s/overlays/staging/`: Patches for staging (e.g., `staging.example.com`, lower resource requests).
- `k8s/overlays/prod/`: Patches for production (e.g., `app.example.com`, high-availability replicas, higher resource requests).

## Security Best Practices
- **Restricted Profile**: Always run containers as non-root users.
- **Privilege Escalation**: Explicitly disable `allowPrivilegeEscalation`.
- **Capabilities**: Drop all default Linux capabilities (`drop: ["ALL"]`).
- **Resources**: Always define CPU and Memory requests and limits to ensure predictable scheduling and prevent "noisy neighbor" issues.
