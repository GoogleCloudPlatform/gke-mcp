#!/bin/bash
set -e

CLUSTER_NAME="$1"
REGION="$2"
PROJECT_ID="$3"

if [ -z "$CLUSTER_NAME" ] || [ -z "$REGION" ] || [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <cluster-name> <region> <project-id>"
    exit 1
fi

echo "Auditing Cluster: $CLUSTER_NAME in $REGION (Project: $PROJECT_ID)"
echo "---------------------------------------------------"

# Get Cluster Details
CLUSTER_JSON=$(gcloud container clusters describe "$CLUSTER_NAME" --region "$REGION" --project "$PROJECT_ID" --format=json)

# Check Workload Identity
WI_CONFIG=$(echo "$CLUSTER_JSON" | jq -r '.workloadIdentityConfig.workloadPool // "DISABLED"')
if [ "$WI_CONFIG" != "DISABLED" ]; then
    echo "[PASS] Workload Identity is ENABLED ($WI_CONFIG)"
else
    echo "[FAIL] Workload Identity is DISABLED"
fi

# Check Network Policy
NETPOL_CONFIG=$(echo "$CLUSTER_JSON" | jq -r '.networkPolicy.enabled // "FALSE"')
if [ "$NETPOL_CONFIG" == "True" ] || [ "$NETPOL_CONFIG" == "true" ]; then
    echo "[PASS] Network Policy is ENABLED"
else
    echo "[FAIL] Network Policy is DISABLED"
fi

# Check Shielded Nodes
SHIELDED_NODES=$(echo "$CLUSTER_JSON" | jq -r '.shieldedNodes.enabled // "FALSE"')
if [ "$SHIELDED_NODES" == "True" ] || [ "$SHIELDED_NODES" == "true" ]; then
    echo "[PASS] Shielded Nodes are ENABLED"
else
    echo "[FAIL] Shielded Nodes are DISABLED"
fi

# Check Binary Authorization
BINAUTH_CONFIG=$(echo "$CLUSTER_JSON" | jq -r '.binaryAuthorization.evaluationMode // "DISABLED"')
if [ "$BINAUTH_CONFIG" != "DISABLED" ]; then
    echo "[PASS] Binary Authorization is ENABLED ($BINAUTH_CONFIG)"
else
    echo "[WARN] Binary Authorization is DISABLED"
fi

# Check Private Cluster
PRIVATE_CLUSTER=$(echo "$CLUSTER_JSON" | jq -r '.privateClusterConfig.enablePrivateNodes // "FALSE"')
if [ "$PRIVATE_CLUSTER" == "True" ] || [ "$PRIVATE_CLUSTER" == "true" ]; then
    echo "[PASS] Private Cluster (Nodes) is ENABLED"
else
    echo "[WARN] Private Cluster (Nodes) is DISABLED"
fi

echo "---------------------------------------------------"
echo "Audit Complete."
