# HEARTBEAT.md - operator

## Cluster Health Patrol
1. Perform a comprehensive health check across all GKE clusters.
2. Check node status (Ready/NotReady) and resource pressure (CPU/Memory/Disk).
3. Identify pods with high restart counts or in CrashLoopBackOff.
4. Look for unschedulable pods due to resource constraints.

## GKE Recommendation Scan
1. Check for new GKE recommendations (e.g., cost optimization, security hardening, or performance improvements).
2. If new actionable recommendations are found, present them with their potential impact and the steps to apply them.

## Critical Error Monitor
1. Scan GKE system logs and application logs for critical error spikes or recurring patterns that might indicate an emerging incident.
2. Focus on K8s events and logs from the `kube-system` namespace.

### Hard Stop
- If anomalies or critical errors are detected, provide a summary of the error logs and suggest investigation paths. Do NOT execute destructive commands or apply major configuration changes without explaining the rationale and seeking explicit human approval.

## Silence rule
- If nothing actionable (the cluster is healthy, no new recommendations, and logs are within normal parameters): reply ONLY `NO_REPLY`.