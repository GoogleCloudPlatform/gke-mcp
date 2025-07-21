# GKE MCP Extension for Gemini CLI

This document provides instructions for an AI agent on how to use the available tools to manage Google Kubernetes Engine (GKE) resources.

## Guiding Principles

*   **Prefer Native Tools:** Always prefer to use the tools provided by this extension (e.g., `list_clusters`, `get_cluster`) instead of shelling out to `gcloud` or `kubectl` for the same functionality. This ensures better-structured data and more reliable execution.
*   **Clarify Ambiguity:** Do not guess or assume values for required parameters like cluster names or locations. If the user's request is ambiguous, ask clarifying questions to confirm the exact resource they intend to interact with.
*   **Use Defaults:** If a `project_id` is not specified by the user, you can use the default value configured in the environment.

## GKE Cost

GKE costs are available from **[GCP Billing Detailed BigQuery Export](https://cloud.google.com/billing/docs/how-to/export-data-bigquery#setup):**. The user will have to provide the full path to their BigQuery table, which inludes their BigQuery dataset name and the table name which contains their Billing Account ID.

These costs can be queried in two ways:
*   **BigQuery CLI:** Using the `bq` command line tool is the preferred way to view the costs, since that can be run locally. If the `bq` CLI is available prefer to use that and offer to run queries for the user.
*   **BigQuery Studio:** If the `bq` CLI is not available, user's can run the query themselves in BigQuery Studio (https://console.cloud.google.com/bigquery).

Some parameters that may be required based on the query:
- Time frame: Assume the last 30 days unless indicated otherwise
- GCP project ID
- GKE cluster location
- GKE cluster name
- Kubernetes namespace (requires [GKE Cost Allocation enabled on the cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/cost-allocations))
- Kubernetes workload type (requires [GKE Cost Allocation enabled on the cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/cost-allocations))
- Kubernetes workload name (requires [GKE Cost Allocation enabled on the cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/cost-allocations))
- Row limit: Assume 10 unless indicated otherwise
- Ordering: Assume ordering by cost descending unless indicated otherwise

When a user asks about a "cluster", a GKE cluster can be uniquely identified with the GCP project ID, the GKE cluster location, and the GKE cluster name.

A GKE workload can be identified by the Kubernetes workload type and Kubernetes workload name. Depending on the scenario, they may want workload costs for a specific cluster and Kubernetes namespace or across all clusters and/or Kubernetes namespaces.

An example BigQuery CLI command for the cost of a single workload in a single cluster is below. All of the above parameters need to be replaced to make it useful.

```sql
bq query --nouse_legacy_sql '
SELECT
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost,
  SUM(cost) AS cost_before_credits,
FROM {{.BQDatasetProjectID}}.{{.BQDatasetName}}.gcp_billing_export_resource_v1_XXXXXX_XXXXXX_XXXXXX AS bqe
WHERE _PARTITIONTIME >= "2025-06-01"
	AND project.id = "sample-project-id"
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AND l.value = "us-central1")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AND l.value = "sample-cluster-name")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "k8s-namespace" AND l.value = "sample-namespace")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "k8s-workload-type" AND l.value = "apps/v1-Deployment")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "k8s-workload-name" AND l.value = "sample-workload-name")
ORDER BY 1 DESC
LIMIT 10
;
'
```

An example BigQuery CLI command for the cost each workload in each cluster is below. All of the above parameters need to be replaced to make it useful.
```sql
bq query --nouse_legacy_sql '
SELECT
  SELECT project.id AS project_id,
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AS cluster_location,
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AS cluster_name,
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-namespace" AS k8s_namespace,
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-workload-type" AS k8s_workload_type,
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-workload-name" AS k8s_workload_name,
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost,
  SUM(cost) AS cost_before_credits,
FROM {{.BQDatasetProjectID}}.{{.BQDatasetName}}.gcp_billing_export_resource_v1_XXXXXX_XXXXXX_XXXXXX AS bqe
WHERE _PARTITIONTIME >= "2025-06-01"
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name")
GROUP BY 1, 2, 3, 4, 5, 6
ORDER BY 7 DESC
LIMIT 10
;
'
```

Checking that the "goog-k8s-cluster-name" label exists scopes the total billing data to just GKE costs.

When using the `bq` CLI, the BQDatasetID needs to use a dot, not a colon, to separate the project and dataset.

The queries can be mixed and adapted to answer a lot of questions about GKE cluster costs.

Many questions the user has about the data produced can be answered by reading the GKE Cost Allocation public documentation at https://cloud.google.com/kubernetes-engine/docs/how-to/cost-allocations. If namespace and workload labels aren't showing up for a particular cluster, make sure the cluster has GKE Cost Allocation enabled.


## GKE Cluster Known Issues

### Objective
To determine if a GKE cluster is impacted by any documented known issues.

### Instructions

1.  **Identify Cluster Versions**: You will need the GKE **control plane (master) version** and the **node pool version(s)** for the cluster you are troubleshooting.

2.  **Consult the Source**: Load https://cloud.google.com/kubernetes-engine/docs/troubleshooting/known-issues into memory. Don't processing URLs in this link.

3.  **Check the Affected Component**: Read the description for each known issue carefully. You must determine if the issue affects the **control plane** or the **node pools**, as the versions for these components can be different.

4.  **Compare and Analyze**: Based on the affected component, compare its version against the specified **"Identified versions"** and **"Fixed versions"** for that issue.

### How to Interpret Version Ranges

A cluster component (either control plane or node pool) is considered **affected** by a known issue if its version is greater than or equal to an **"Identified version"** and less than the corresponding **"Fixed version"**.

* **Rule**: A component is affected if `identified_version <= component_version < fixed_version`.

* **Example**:
    * A known issue lists the following versions and specifies it affects **node pools**:
        * **Identified versions**: `1.28`, `1.29`
        * **Fixed versions**: `1.28.7-gke.1026000`, `1.29.2-gke.1060000`
    * **Conclusion**: A node pool is affected if its version falls into either of these ranges:
        * Between `1.28` (inclusive) and `1.28.7-gke.1026000` (exclusive).
        * Between `1.29` (inclusive) and `1.29.2-gke.1060000` (exclusive).
