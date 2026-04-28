---
name: gke-ai-troubleshooting-skill-creation-guide
description: >
  Expert instructions for building high-quality GKE troubleshooting skills.
  Codifies Step 0 context rules, zero-hallucination signatures, and explicit
  LQL/PromQL query requirements.
---

# Troubleshooting Skill Creation Guide

Use this guide to build high-quality troubleshooting skills that enable AI
agents to diagnose complex failures in GKE workloads.

## 🏗️ Skill Structure Standard

### Mandatory Components

1. **`SKILL.md`**: The core diagnostic and resolution workflow.
2. **`README.md`**: Public-facing overview and "When to use" guide.
3. **`references/failure_signatures.md`**: Authentic log/metric signatures.
4. **`scripts/validate_queries.sh`**: Automatic syntax validator for all
   queries.
5. **`TEST.md`**: Manual verification plan for humans.
6. **`EVAL.textproto`**: Evaluation suite for performance tracking.

### Optional Components

1. **`BUILD`**: Build definition.

## 🏷️ Naming Conventions

- **Directory Name**: MUST be `kebab-case` (e.g.,
  `gke-ai-troubleshooting-tpu-vbar-oom`).
- **Skill Name**: MUST match the directory name.

## 🔍 Diagnostic Workflow Standards

### Step 0: Mandatory Context

Every skill MUST begin with a "Step 0" to acquire necessary context.

- **Mandatory Fields**: `<project_id>`, `<location>`, `<cluster_name>`,
  `<timestamp>`.
- **Optional/Case-by-Case Fields**: `<node_name>`, `<workload_name>`,
  `<workload_namespace>`, `<nodepool_name>`.
- **Time Rule**: Reject relative time (e.g., "5 minutes ago"). Calculate a
  window of `[T - 30m]` to `[T + 30m]`.

### Diagnostic Steps

- **Explicit Queries**: Every step MUST provide a ready-to-use **Cloud Logging
  (LQL)** or **Cloud Monitoring (PromQL)** query.
- **Placeholder Syntax**: Use **angle brackets** like `<project_id>` instead
  of curly braces for placeholders to avoid template resolution errors.
- **Risk Categorization**: Label every step as **[Low Risk]** (Read-only) or
  **[High Risk]** (Mutative/Destructive).
- **Automation**: Specify if the agent should proceed automatically or wait
  for user confirmation.

## 🛠️ Accuracy & Validation

### Zero Hallucination

- Never synthesize example logs or metrics.
- Source signatures from real incidents and anonymize where necessary.
- **DO NOT EXTRAPOLATE**: Only include steps and queries that were verified
  in the source conversation.

### Security & Privacy

- **No Raw Dumps**: Do not instruct the agent to dump raw logs into shared
  spaces (bugs, chat).
- **Signal Only**: Instruct the agent to summarize findings and report only
  high-signal information (e.g., "Found specific error pattern X on node Y").

### Automated Validation

- Every skill MUST include a script (at `scripts/validate_queries.sh`) that
  uses `query_logs` or `gcloud logging read ... --limit=1` to verify its LQL
  queries.

## 📋 Best Practices

- **Conciseness**: Keep instructions lean. Focus on "what to do" and "how to
  verify".
- **Public Ready**: Remove all internal notes, personal bookmarks, or
  project-specific jargon.
- **Error Signatures**: Explicitly link to `references/failure_signatures.md`
  in relevant diagnostic steps.
