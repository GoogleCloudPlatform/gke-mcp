# Template: Generated GKE Skill

All generated skills must strictly follow this structure.

**Guidelines for generated skills:**

-   **Cheatsheet Philosophy**: Keep it terse, opinionated, and focused on
    "gotchas" and exact command patterns. Avoid long paragraphs of explanation.
-   **Explain the Why**: For remediation steps, briefly explain *why* the
    command is necessary, not just what it does.
-   **Public Only**: Ensure NO internal codenames, acronyms, or internal system
    references are included.
-   **Intent Matching**: Direct executing agents to match user intent
    (explaining/investigating vs. actually applying a fix).

--------------------------------------------------------------------------------

name: [issue-name-kebab-case] description: >-

## Expert diagnosis and remediation for [Issue Name].

Triggered by [Observable Symptom].

## [Issue Name] Expert

### Symptoms

-   [Symptom 1]
-   [Symptom 2]
-   [Specific Error Message/Code]

### Prerequisites

List any specific permissions, tools, or cluster state required before
proceeding.

-   [e.g., Cluster admin permissions]
-   [e.g., `kubectl` version >=1.x]

### User Intent & Execution Rules

-   **Explain/Investigate**: If requested, only explain or diagnose. Do not
    execute modifying commands.
-   **Fix/Resolve**: If asked to fix the issue, actually execute the commands to
    apply the fix rather than just listing them.

### Diagnosis (Pre-flight Checks)

Instructions for the agent to run specific read-only commands to verify the
cluster is in the expected state.

1.  Run `kubectl get pods -n [namespace]` to check status.
2.  Run `kubectl describe [resource] [name]` to look for events.
3.  Check logs using `kubectl logs [pod]`.

### Fix (Remediation)

Step-by-step raw `kubectl` or `gcloud` commands to perform the necessary
actions. *Note: Every command block MUST be preceded by a brief description of
its impact and a short rationale explaining why it is needed.*

1.  **[Action Description]**:

   ```bash
   kubectl [command]
   ```

2.  **[Destructive Action Description]**:
   *CRITICAL: Ask the user for permission before executing the following
    command.*

   ```bash
   kubectl delete [resource] [name]
   ```

### Verification

Mandatory commands to run to confirm the fix is effective.

1.  Run `kubectl get [resource]` and verify [Expected State].
2.  Run `kubectl describe` and confirm no further errors.

### Rollback / Cleanup

Steps to undo the changes if the fix fails or causes issues.

1.  **[Rollback Action]**: `kubectl [command to undo]`

### Technical Explanation

A concise explanation of the root cause and why the fix works. [Brief technical
details...]

### Known Limitations and Edge Cases

List any situations where this skill might not work or requires special
attention.

-   [e.g., Does not work on Autopilot clusters]
-   [e.g., May cause brief downtime]
