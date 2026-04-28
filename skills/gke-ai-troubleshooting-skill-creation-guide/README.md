# GKE AI Troubleshooting Skill Creation Guide

This guide provides a standardized framework and utility for building
high-quality, production-ready troubleshooting skills for GKE.

## Purpose

Codifying best practices ensures that diagnostic skills are:

1.  **Predictable**: Consistent Step 0 context acquisition.
2.  **Accurate**: Built on authentic failure signatures, never hallucinations.
3.  **Actionable**: Every step provides explicit LQL or PromQL queries.
4.  **Verifiable**: Includes built-in syntax validation.

## Getting Started

To create a new skill following these practices, use the provided bootstrap
script:

```bash
./scripts/bootstrap_skill.sh <your-skill-name>
```

## Core Standards

Refer to the **[SKILL.md](./SKILL.md)** in this directory for the detailed
"expert instructions" on how to build and validate a new skill.
