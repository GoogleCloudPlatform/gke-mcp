---
name: mvp-to-prod
description: 'Standard workflow for promoting applications from MVP to production-grade deployments on GKE. Use when: (1) Setting up staging and production environments, (2) Implementing PR-based rollout gates, or (3) Configuring environment-specific parameters via Kustomize.'
---

# GKE MVP-to-Production Workflow

## Overview

This skill provides a standardized, production-grade workflow for managing application lifecycles on GKE, from initial build to production rollout.

## Workflow

1.  **Environment Setup**: Maintain separate `staging` and `prod` namespaces for isolation.
2.  **GitOps & Review**: Ensure every change goes through a Pull Request. No direct commits to the `main` branch.
3.  **Build**: Use Cloud Build to create versioned, immutable container images.
4.  **Staging Deployment**: Use `kubectl apply -k k8s/overlays/staging` to deploy and validate changes.
5.  **Quality Gate**: Monitor for stability in `staging` for a defined period (e.g., 6 hours) before promotion.
6.  **Production Promotion**: Roll out to `prod` using `kubectl apply -k k8s/overlays/prod` within a maintenance window.

## How to use

- "Set up a new staging environment for my application."
- "Promote the current stable version from staging to production."
- "Configure Kustomize overlays for separate staging and prod domains."
- "Review my current rollout strategy against the mvp-to-prod best practices."
