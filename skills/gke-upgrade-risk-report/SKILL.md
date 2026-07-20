---
name: gke-upgrade-risk-report
description: Generate a comprehensive upgrade risk report for GKE cluster, analyzing upgrade changes and cluster state.
---

# GKE Upgrade Risk Report

## Purpose

This skill guides the generation of a comprehensive upgrade risk report for a GKE cluster. It helps identify potential issues, breaking changes, and deprecated APIs introduced between the cluster's current version and a target upgrade version, providing actionable verification and mitigation recommendations.

## When to Use

- **Goal:** Identify risks before upgrading a GKE cluster to a new version.
- **Context:** User asks about safety of GKE cluster upgrade.

## Prerequisites

- Access to the GKE cluster via `gcloud` and `kubectl`.

## Workflow

### 1. Input Verification

Ensure you have the following parameters:
- **Project ID**
- **Cluster Location** (Zone or Region)
- **Cluster Name**
- **Target Version**

> [!IMPORTANT]
> **Handling Missing Project ID:**
> If Project ID is not provided, use `gcloud config get-value project` to get the current project ID. If it's not set, ask user to provide it.

> [!IMPORTANT]
> **Handling Missing Cluster Name and Location:**
> If Cluster Name and Location are not provided, use `gcloud container clusters list --project <PROJECT_ID> --format 'table(name,location)'` to get the current cluster names and locations, provide it to user to select from.

> [!IMPORTANT]
> **Handling Missing Target Version:**
> If the **Target Version** is not provided:
> 1. State that the target version is required.
> 2. Fetch available GKE versions for the location:
>    ```bash
>    gcloud container get-server-config --zone=<ZONE> # or --region=<REGION>
>    ```
> 3. Filter the list to show only versions **newer** than the cluster's current version and compatible with the release channel.
> 4. Present these options to help select a target.

### 2. Information Gathering

Gather data to support the analysis:
- **Cluster Details:** Use `gcloud container clusters describe` to get the cluster state (including current control plane version, node pool versions, release channel, etc).
- **In-Cluster Resources:** Use `kubectl` (e.g., `kubectl api-resources`, `kubectl get`) to inspect workloads and APIs in use to match against risky upgrade changes.
- **Changelogs & Release Notes:** Fetch Kubernetes changelogs using the `get_k8s_changelog` tool from GKE MCP server and fetch GKE release notes using `get_gke_release_notes` tool from GKE MCP server. Extract data for the version range:
  - **Minor Versions:** Analyze changes from the current minor version up to and including the target minor version.
  - **Patch Versions:** Analyze changes for every patch version between the current version (exclusive) and the target version (inclusive).
  - **GKE Versions:** Analyze GKE release notes for all versions between the current version (exclusive) and the target version (inclusive).

### 3. Risk Identification & Analysis

Analyze upgrade changes with a focus on:
- **API Deprecations & Removals:** Specifically those affecting resources currently running in the cluster.
- **Breaking Changes:** Significant behavioral changes in stable features.
- **Default Configuration Changes:** Modifications that may alter workload behavior.
- **Interaction Risks:** Disruptive interactions between new and old features.

> [!IMPORTANT]
> **Filter Rule:**
> In the final report:
> - **Only include risks that require mitigation actions**. Exclude new features, performance improvements, or risks that cannot be mitigated prior to the upgrade.
> - **Do not include** risks that are not relevant for the cluster and its workloads (analyse only user-managed workloads).

### 4. Report Format

Present the risks in a list ordered by severity. Keep each risk item concise but comprehensive. Use the following structured markdown for **each** risk item:

```markdown
# Short Risk Title

## Description
(Detailed description of the change and the potential risk it introduces for this specific upgrade)

## Verification Recommendations
(Clear, actionable steps or commands to check if the cluster is affected. Include `kubectl` or `gcloud` examples and documentation links.)

## Mitigation Recommendations
(Clear, actionable steps, configuration changes, or code adjustments to mitigate the risk BEFORE the upgrade. Provide examples and links to documentation.)
```
