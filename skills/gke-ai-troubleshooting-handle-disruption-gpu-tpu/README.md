# Handle Disruption on GPUs and TPUs Troubleshooting Skill

## Purpose

This skill equips the troubleshooting agent with the ability to diagnose and remediate unexpected node terminations affecting GPU and TPU workloads on GKE.

## When to use

This skill should be triggered when a user reports:

- GPU or TPU workloads crashing or restarting unexpectedly.
- Nodes with accelerators shutting down without apparent cause.
- Suspected host maintenance events.

## Mechanism

The skill uses:

1. **Cloud Monitoring** to check `kubernetes_io:node_interruption_count` for `HW/SW Maintenance`.
2. **Cloud Logging** to check for `cloud.google.com/active-node-maintenance` and the `impending-node-termination:NoSchedule` taint.
3. **kubectl** to predict upcoming maintenance using the `cloud.google.com/scheduled-maintenance-time` label.
