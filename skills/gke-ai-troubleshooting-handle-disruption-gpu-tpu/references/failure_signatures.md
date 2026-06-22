# Failure Signatures for GPU/TPU Host Maintenance

## Cloud Logging Signatures

- `cloud.google.com/active-node-maintenance: ONGOING` (Log payload indicating maintenance has started and workloads are actively being stopped).
- `cloud.google.com/impending-node-termination:NoSchedule` (Taint applied to the node to prevent new pods from scheduling).

## Cloud Monitoring Signatures

- `kubernetes_io:node_interruption_count` metric > 0.
- `interruption_reason="HW/SW Maintenance"` label on the metric.

## Kubectl Label Signatures

- `cloud.google.com/scheduled-maintenance-time` (Node label containing the UNIX epoch timestamp of upcoming scheduled maintenance).
