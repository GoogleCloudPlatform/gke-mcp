---
name: mvp-to-prod
description: 'Standardized workflow for promoting applications from MVP to production-grade deployments on GKE. Use when: (1) Designing multi-environment GKE architectures, (2) Implementing PR-based rollout gates, or (3) Configuring environment-specific parameters via Kustomize overlays.'
---

# GKE MVP-to-Production Workflow

## Overview

This skill provides a comprehensive, production-grade workflow for managing application lifecycles on GKE. It transitions applications from an initial MVP state to a reliable, monitored, and multi-environment deployment using GitOps principles and GKE best practices.

## Workflow

### 1. Environment & Infrastructure Isolation
- **Namespacing**: Maintain strictly separate namespaces for `staging` and `prod` to ensure isolation of resources and security contexts.
- **Resource Definition**: Always define explicit CPU and Memory requests and limits for every container to ensure predictable scheduling.
- **Security Hardening**: Enforce the **Restricted Pod Security Standard** (e.g., `runAsNonRoot: true`, `allowPrivilegeEscalation: false`, and `capabilities.drop: ["ALL"]`).

### 2. GitOps & PR-Based Lifecycle
- **Version Control**: All code, Dockerfiles, and K8s manifests must reside in a version-controlled repository.
- **Mandatory Review**: Every change (even minor manifest updates) must go through a Pull Request. Direct commits to the primary branch are forbidden.
- **Build Automation**: Use **Cloud Build** to create versioned, immutable container images. Avoid using the `latest` tag; use semantic versioning or commit-SHAs.

### 3. Staging as a Quality Gate
- **Kustomize Overlays**: Use a `k8s/base` directory for generic resources and environment-specific overlays (e.g., `k8s/overlays/staging`) to handle parameters like hostnames or replica counts.
- **Deployment**: Apply changes to the staging namespace: `kubectl apply -k k8s/overlays/staging`.
- **The Stability Crucible**: Monitor the application in staging for a mandatory period (e.g., 6 consecutive hours). 
  - Application-level crashes (e.g., `CrashLoopBackOff`, `OOMKilled`) reset the stability clock.
  - Normal infrastructure churn (e.g., GKE Spot node preemptions) should be filtered out and not reset the clock if the app recovers automatically.

### 4. Controlled Production Promotion
- **Promotion Rule**: Only promote the exact image version and configuration that has successfully passed the staging stability crucible.
- **Rollout**: Apply the production overlay: `kubectl apply -k k8s/overlays/prod`.
- **Maintenance Windows**: Execute production deployments only within defined maintenance windows (e.g., 1 AM - 4 AM local time) to minimize business risk.

## Examples & Scenarios

### Example 1: Updating an Application Feature
1. **Developer**: Submits a PR with code changes and an updated image tag in `k8s/base/kustomization.yaml`.
2. **CI/CD**: Cloud Build triggers, builds the image, and pushes to Artifact Registry.
3. **Agent**: Detects the merge, applies the update to `staging`, and starts the 6-hour stability timer.
4. **Agent**: After 6 hours of zero-crash stability, the agent alerts the operator and prepares the production rollout.

### Example 2: Handling a Staging Failure
1. **Agent**: Deploys a new version to `staging`.
2. **Issue**: A pod crashes due to a missing environment variable.
3. **Action**: The agent detects the `CrashLoopBackOff`, resets the stability timer, and notifies the developer.
4. **Resolution**: The fix is submitted via a new PR; the promotion cycle restarts.

## How to use

- "Configure a new application with a base manifest and two Kustomize overlays for staging and production."
- "Start monitoring the stability of my current deployment in staging."
- "What is the status of the staging-to-prod stability crucible?"
- "Verify that my staging and production manifests comply with the GKE restricted security standards."
- "Promote the current staging deployment to production within the next maintenance window."
