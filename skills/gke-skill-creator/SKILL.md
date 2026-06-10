---
name: gke-skill-creator
description: >-
  Dynamically generates specialized GKE skills for complex troubleshooting,
  operational workflows, architectural setup, or performance/cost optimization.
  Trigger this skill whenever the user faces a novel or non-obvious GKE
  challenge, needs custom cluster management workflows, or standard agent
  capabilities fall short, even if they don't explicitly ask to create a skill.
---

# GKE Skill Creator

## Overview

This meta-skill acts as an interactive assistant to diagnose novel or complex
Google Kubernetes Engine (GKE) issues and dynamically generate a specialized
troubleshooting skill tailored to the exact problem discovered.

## Core Mandate: Public-Facing Output

During the investigation and research phase, you should leverage official GKE
documentation, public Kubernetes issue trackers, and trusted SRE resources to
deeply understand the issue.

**CRITICAL:** The _generated skill_ must be strictly public-facing. It MUST NOT
contain any internal codenames, acronyms, or references to internal Google
systems. All remediation logic must be expressed in terms of standard, public
tools like `kubectl` and `gcloud`. This ensures the skill is usable by external
GKE customers.

## Workflow

### 1. Capture Intent & Investigate

- Ask the user for specific GKE symptoms (e.g., Pods stuck in
  `CrashLoopBackOff`, Service 503 errors, Node pressure).
- Use **read-only** tools to pull live context:
  - `kubectl get <resource> -o yaml`
  - `kubectl describe <resource>`
  - `kubectl logs <pod> --tail=100`
  - `kubectl get events --sort-by='.lastTimestamp'`
  - `gcloud container clusters describe <cluster>`

### 2. Perform Deep Research

- Before drafting the skill, perform deep research on the identified topic.
- Understand default argument values, potential side effects, and best
  practices for the commands you plan to include.
- Use trusted public sources to ensure accuracy.
- Use web search to find specific technical details from official
  documentation and trusted sources (e.g., Google Cloud, GKE, Kubernetes),
  such as:
  - Exact error message matches and their documented causes.
  - Command references for `kubectl` or `gcloud` to verify syntax and flags.
  - Known issues, limitations, or version-specific caveats.
  - Official troubleshooting workflows and decision trees.

### 3. Draft the Skill

- Use the
  [generated_skill_skeleton.md](references/generated_skill_skeleton.md)
  template.
- **Cheat sheet Philosophy**: Write the `SKILL.md` to be terse and opinionated.
  Focus on "gotchas", exact command patterns, and specific pitfalls rather
  than long explanations.
- **Progressive Disclosure**: For complex issues, avoid a monolithic
  `SKILL.md`. Suggest breaking down long lists of commands or log analysis
  patterns into separate reference artifacts in `references/`.
- **Make Descriptions Pushy**: Ensure the generated skill's description
  explicitly states when it should be used, covering variations of the
  problem.
- **Fallback Remediation**: If the primary method depends on specific
  high-level tools (e.g., MCP tools, API integrations) or environment
  configurations that might fail, include standard CLI fallback alternatives
  (like raw `kubectl` or `gcloud` commands) to achieve the same result.
- **Match User Intent**: Instruct executing agents to match the user's intent:
  explain/investigate if requested, or execute remediation if asked to fix the
  issue.
- Ensure the draft contains:
  - Precise Symptoms.
  - User Intent & Execution Rules.
  - Step-by-step Diagnosis commands.
  - Remediation (Fix) commands with impact descriptions.
  - Verification steps.
  - Technical Explanation (explaining _why_ the fix works).

### 4. Review & Iterate

- Present the proposed diagnosis and the draft commands to the user.
- **Human-in-the-loop:** The user must approve the logic before finalization.
- Incorporate any user feedback or constraints.

### 5. Finalize & Handoff

- Once approved, provide the finalized `SKILL.md` content directly in the
  chat.
- **Handoff:** Instruct the user that they can use this diagnosis and
  remediation logic directly in their current session, or save it to a local
  file for reference.
- Clarify that this process is for generating dynamic, session-specific skills
  and is distinct from adding permanent skills to the GKE-MCP codebase.

## Safety Guardrails

- **Read-Only Discovery:** Never execute modifying commands during the
  investigation phase.
- **Destructive Actions:** Generated skills MUST instruct the agent to seek
  explicit human confirmation before running destructive commands (e.g.,
  `kubectl delete`, `gcloud container clusters update`).
- **Impact Description:** Every remediation command must have a clear
  explanation of what it does.

## References

- [Generated Skill Template](references/generated_skill_skeleton.md)
