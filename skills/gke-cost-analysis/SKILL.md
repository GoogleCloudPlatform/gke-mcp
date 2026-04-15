---
name: gke-cost-analysis
description: Answer natural language questions about GKE-related costs by leveraging BigQuery export and cost allocation data.
---

# GKE Cost Analysis

This skill provides guidance on answering natural language questions about GKE-related costs, optimization, or billing.

## Overview

When users ask about GKE costs (e.g., "What are my costs across projects?", "What's my most expensive namespace?"), use this skill to provide a structured and expert response.

## Instructions

When handling a cost-related question:

1. **Provide a Direct Answer**: Address the specific cost question or optimization request.
2. **Explain BigQuery Integration**: Explain how to use BigQuery for cost analysis if relevant. Mention that GKE costs come from GCP Billing Detailed BigQuery Export.
3. **Check Cost Allocation**: Mention that GKE Cost Allocation must be enabled for namespace and workload-level cost data.
4. **Provide Actionable Steps**: Provide concrete next steps or commands when possible. Prefer BigQuery CLI (`bq`) over BigQuery Studio when available.
5. **Reference Resources**: Point to relevant GCP documentation or console links.

## Key Points to Remember

- **Data Source**: GKE costs come from GCP Billing Detailed BigQuery Export. The user must provide the full path to their BigQuery table (dataset name and table name containing Billing Account ID).
- **Granularity**: GKE Cost Allocation must be enabled for namespace and workload-level cost data.
- **Tools**: BigQuery CLI (`bq`) is preferred. When writing Standard SQL queries, use a dot (`.`) instead of a colon (`:`) to separate the project ID and dataset name.
- **Defaults**: Assume last 30 days, row limit 10, ordering by cost descending, unless specified otherwise.

## Example BigQuery Queries

Use these queries as templates to answer questions. All parameters (dataset, table, project, cluster, etc.) need to be replaced.

### Cost of a Single Workload in a Single Cluster

```sql
bq query --nouse_legacy_sql '
SELECT
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost,
  SUM(cost) AS cost_before_credits
FROM {{.BQDatasetProjectID}}.{{.BQDatasetName}}.gcp_billing_export_resource_v1_XXXXXX_XXXXXX_XXXXXX AS bqe
WHERE _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
  AND project.id = "sample-project-id"
  AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AND l.value = "us-central1")
  AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AND l.value = "sample-cluster-name")
  AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "k8s-namespace" AND l.value = "sample-namespace")
  AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "k8s-workload-type" AND l.value = "apps/v1-Deployment")
  AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "k8s-workload-name" AND l.value = "sample-workload-name")
;
'
```

### Cost of Each Workload in Each Cluster

```sql
bq query --nouse_legacy_sql '
SELECT
  project.id AS project_id,
  (SELECT l.value FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" LIMIT 1) AS cluster_location,
  (SELECT l.value FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" LIMIT 1) AS cluster_name,
  (SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-namespace" LIMIT 1) AS k8s_namespace,
  (SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-workload-type" LIMIT 1) AS k8s_workload_type,
  (SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-workload-name" LIMIT 1) AS k8s_workload_name,
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost,
  SUM(cost) AS cost_before_credits
FROM {{.BQDatasetProjectID}}.{{.BQDatasetName}}.gcp_billing_export_resource_v1_XXXXXX_XXXXXX_XXXXXX AS bqe
WHERE _PARTITIONTIME >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
  AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name")
GROUP BY 1, 2, 3, 4, 5, 6
ORDER BY 7 DESC
LIMIT 10
;
'
```

Note: Checking that the "goog-k8s-cluster-name" label exists scopes the total billing data to just GKE costs.
