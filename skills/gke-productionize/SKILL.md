---
name: gke-productionize
description: Assists in preparing applications and clusters on GKE for production.
---

# GKE Productionize Skill

This skill acts as a high-level orchestrator for preparing a GKE cluster and its workloads for production readiness.

> [!IMPORTANT]
> This is a **meta-skill** or **orchestrator skill**. You are expected to invoke and run many other specialized skills listed in this document as part of the overall productionization process. Do not attempt to implement all production readiness features directly within this skill; instead, use this skill to assess the environment and then delegate to the specific skills for each domain.

## Scope

This skill is adaptable to:

- A single application (already on Kubernetes or not).
- A set of applications.
- A target cluster.

## Workflow

### 1. Discovery Phase

Before making recommendations, discover the current state of the environment.

#### Cluster Discovery

Run these commands to understand the cluster setup:

- Check cluster details: `gcloud container clusters describe <cluster-name> --location <location> --project <project>`
- Check for Autopilot vs Standard: Look for `autopilot: true` in the describe output.
- Check release channel: Look for `releaseChannel`.

#### Workload Discovery

If a specific application is targeted, discover its configuration:

- Get deployment/statefulset details: `kubectl get deployment <app-name> -n <namespace> -o yaml`
- Check for resource requests and limits.
- Check for liveness, readiness, and startup probes.
- Check for HPA: `kubectl get hpa -n <namespace>`
- Check for PDB: `kubectl get pdb -n <namespace>`
- Check for NetworkPolicies: `kubectl get networkpolicy -n <namespace>`

### 2. Production Readiness Assessment

**Before implementation, you MUST view the `SKILL.md` file for each relevant specialized skill listed below and incorporate its guidance into your assessment and plan. Failure to do so will result in a non-compliant production configuration.**

#### A. App Onboarding (Pre-Kubernetes)

If the application is not yet running on GKE, delegate to the [gke-app-onboarding](../gke-app-onboarding/SKILL.md) skill for containerization and initial deployment.

#### B. Scalability & Resource Management

Ensure workloads have appropriate resources and autoscaling.

- **Action**: You MUST activate [gke-workload-scaling](../gke-workload-scaling/SKILL.md) for configuring HPA, VPA, and resource limits.

#### C. Observability

Ensure adequate logging and monitoring are in place.

- **Action**: You MUST activate [gke-observability](../gke-observability/SKILL.md) for setting up Cloud Logging, Monitoring, and Managed Prometheus.

#### D. Reliability

Ensure high availability and graceful degradation.

- **Action**: You MUST activate [gke-reliability](../gke-reliability/SKILL.md) for configuring regional clusters, PDBs, and health probes.

#### E. Security

Harden the cluster and workloads.

- **Action**: You MUST activate [gke-workload-security](../gke-workload-security/SKILL.md) for Workload Identity, Network Policies, and Shielded Nodes.

#### F. Backup & Disaster Recovery

Ensure stateful data is protected.

- **Action**: You MUST activate [gke-backup-dr](../gke-backup-dr/SKILL.md) for configuring Backup for GKE and restore procedures.

#### G. Edge Security & Ingress

Secure external access.

- **Action**: You MUST activate [gke-networking-edge](../gke-networking-edge/SKILL.md) for Gateway API, Ingress, and Cloud Armor.

#### H. Cost Optimization

Ensure efficient use of resources.

- **Action**: You MUST activate [gke-cost-optimization](../gke-cost-optimization/SKILL.md) for strategies on rightsizing, quotas, and Spot VMs.

### 3. Production Readiness Scoring

After the assessment, provide a summary report with a RAG (Red, Amber, Green) status for each area and an overall readiness score. This helps prioritize remediation efforts.

## Adaptability Guidelines

- **Single App**: Focus on Health Probes, HPA, Resource Limits, PDB, and Workload Identity for that specific app.
- **Cluster Wide**: Focus on Cluster Autoscaler, Multi-zonal setup, Release Channels, Maintenance Windows, and default Network Policies.
- **Proactive Execution**: Proactively execute relevant skills (e.g., observability, security, scaling, reliability) to assess and propose improvements, seeking user confirmation before applying state-changing implementations.
