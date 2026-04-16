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

Go through these areas and assess readiness. **You should proactively execute the relevant specialized skills to perform the detailed assessment and implementation for each domain.**

#### A. App Onboarding (Pre-Kubernetes)

If the application is not yet running on GKE, delegate to the [gke-app-onboarding](../gke-app-onboarding/SKILL.md) skill for containerization and initial deployment.

#### B. Scalability & Resource Management

Ensure workloads have appropriate resources and autoscaling.

- **Action**: Delegate to [gke-workload-scaling](../gke-workload-scaling/SKILL.md) for configuring HPA, VPA, and resource limits.

#### C. Observability

Ensure adequate logging and monitoring are in place.

- **Action**: Delegate to [gke-observability](../gke-observability/SKILL.md) for setting up Cloud Logging, Monitoring, and Managed Prometheus.

#### D. Reliability

Ensure high availability and graceful degradation.

- **Action**: Delegate to [gke-reliability](../gke-reliability/SKILL.md) for configuring regional clusters, PDBs, and health probes.

#### E. Security

Harden the cluster and workloads.

- **Action**: Delegate to [gke-workload-security](../gke-workload-security/SKILL.md) for Workload Identity, Network Policies, and Shielded Nodes.

#### F. Backup & Disaster Recovery

Ensure stateful data is protected.

- **Action**: Delegate to [gke-backup-dr](../gke-backup-dr/SKILL.md) for configuring Backup for GKE and restore procedures.

#### G. Edge Security & Ingress

Secure external access.

- **Action**: Delegate to [gke-networking-edge](../gke-networking-edge/SKILL.md) for Gateway API, Ingress, and Cloud Armor.

#### H. Deployment & GitOps

Adopt modern deployment and configuration management practices.

- **Action**: Delegate to [gke-cluster-lifecycle](../gke-cluster-lifecycle/SKILL.md) for **Config Sync** (GitOps-driven configuration management) or [gke-gitops-delivery](../gke-gitops-delivery/SKILL.md) for application CD (e.g., Cloud Deploy).

#### I. Image Streaming (Advanced Optimization)

Reduce startup times for large container images by streaming image data as needed.

- **Action**: Enable Image Streaming on the cluster/node pool.
  - **Standard**: `gcloud container node-pools update <pool-name> --cluster <cluster-name> --enable-image-streaming`
  - **Autopilot**: Automatically enabled for eligible images (on Container Registry or Artifact Registry).

#### J. Cost Optimization

Ensure efficient use of resources.

- **Action**: Delegate to [gke-cost-optimization](../gke-cost-optimization/SKILL.md) for strategies on rightsizing, quotas, and Spot VMs.

### 3. Production Readiness Scoring

After the assessment, provide a summary report with a RAG (Red, Amber, Green) status for each area and an overall readiness score. This helps prioritize remediation efforts.

## Production Readiness Checklist

Before going live, verify the following:

- [ ] **Cluster**: Regional cluster, Release Channel (Stable/Regular), Maintenance Window configured.
- [ ] **Workloads**: Resource requests/limits set, Liveness/Readiness probes active, HPA/VPA enabled.
- [ ] **Security**: Workload Identity used, Network Policies enforced, Private Cluster enabled.
- [ ] **Reliability**: PDBs configured, Pods spread across zones (Topology Spread).
- [ ] **Observability**: Logging enabled, Managed Prometheus active, Dashboards/Alerts set up.
- [ ] **Storage**: CSI drivers used, Volume expansion enabled, Backups configured.

## Adaptability Guidelines

- **Single App**: Focus on Health Probes, HPA, Resource Limits, PDB, and Workload Identity for that specific app.
- **Cluster Wide**: Focus on Cluster Autoscaler, Multi-zonal setup, Release Channels, Maintenance Windows, and default Network Policies.
- **Proactive Execution**: Proactively execute relevant skills (e.g., observability, security, scaling, reliability) to assess and propose improvements, seeking user confirmation before applying state-changing implementations.
