# Skill: GKE AI Troubleshooting - TPU Connection Failure & VBAR OOM

This skill provides an automated diagnostic and resolution workflow for
`vbar_control_agent` crashes on TPU v6e nodes.

## What is this issue

On TPU v6e nodes, a race condition can occur during TPU device resets if
frequent metrics collection (e.g. every 3s) is active. This can lead to memory
corruption in the `vbar_control_agent` process, resulting in segmentation faults
or cgroup OOMs.

## When to use this skill

- TPU slice initialization fails with "connection refused" to VBAR.
- `vbar_control_agent` crashes and restarts on TPU nodes.
- `tpu-device-plugin` logs show "checksum didn't match with the metrics data".

## Components

- `SKILL.md`: Main instruction set for AI agents.
- `references/failure_signatures.md`: Example failure logs for pattern matching calibration.
- `scripts/validate_queries.sh`: Syntax validator for diagnostic queries.
- `TEST.md`: Manual verification plan.
- `EVAL.textproto`: Evaluation data for performance measurement.

## Maintenance

Filters in `SKILL.md` should be periodically validated using
`scripts/validate_queries.sh` to ensure compatibility with Cloud Logging
schema updates.
