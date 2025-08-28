# GKE MCP Extension for Gemini CLI

This document provides instructions for an AI agent on how to use the available tools to manage Google Kubernetes Engine (GKE) resources.

## Guiding Principles

- **Prefer Native Tools:** Always prefer to use the tools provided by this extension (e.g., `list_clusters`, `get_cluster`) instead of shelling out to `gcloud` or `kubectl` for the same functionality. This ensures better-structured data and more reliable execution.
- **Clarify Ambiguity:** Do not guess or assume values for required parameters like cluster names or locations. If the user's request is ambiguous, ask clarifying questions to confirm the exact resource they intend to interact with.
- **Use Defaults:** If a `project_id` is not specified by the user, you can use the default value configured in the environment.
- **Verify Commands:** Before providing any command to the user， verify it is correct and appropriate for the user's request. You can search online or refer to [gcloud documentation](https://cloud.google.com/sdk/gcloud).

## Authentication

Some MCP tools required [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). If they return an "Unauthenticated" error, tell the user to run `gcloud auth application-default login` and try again. This is an interactive command and must be run manually outside the AI.

## GKE Logs

- When searching for GKE logs, always use the `query_logs` tool to fetch them. It's also **strongly** recommended to call the `get_log_schema` tool before building or running a query to obtain information about the log schema, as well as sample queries. This information is useful when building Cloud Logging LQL queries.

- When using time ranges, make sure you check the current time and date if the range is relative to the current time or date.

- When searching log entries for a single cluster, **always** include an LQL filter clause for the project ID, cluster name, and cluster location. Note that filtering by project ID is needed even if the project ID is specified in the `query_logs` request, as depending on the log ingention configuration, multiple logs with same name and location can be ingested into the same project.

- If you need help understanding LQL syntax, consider fetching [Logging query language](https://cloud.google.com/logging/docs/view/logging-query-language) to learn more about it.

## GKE Monitoring

When users ask a question about the Monitoring or monitored resource types, the following instructions could be applied:

- Please use the tool `list_monitored_resource_descriptors` to get all monitored resource descriptors
- After getting all the monitored resource, if the user ask for GKE specific ones, please filter the output and only include the GKE related ones
  \*\* Full GKE related monitored resources are the one contains `gke` or `k8s` or `container.googleapis.com`

## GKE Cost

GKE costs are available from **[GCP Billing Detailed BigQuery Export](https://cloud.google.com/billing/docs/how-to/export-data-bigquery#setup):**. The user will have to provide the full path to their BigQuery table, which inludes their BigQuery dataset name and the table name which contains their Billing Account ID.

These costs can be queried in two ways:

- **BigQuery CLI:** Using the `bq` command-line tool is the preferred way to view the costs, since that can be run locally. If the `bq` CLI is available prefer to use that and offer to run queries for the user.
- **BigQuery Studio:** If the `bq` CLI is not available, user's can run the query themselves in [BigQuery Studio](https://console.cloud.google.com/bigquery).

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

Many questions the user has about the data produced can be answered by reading the [GKE Cost Allocation public documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/cost-allocations). If namespace and workload labels aren't showing up for a particular cluster, make sure the cluster has GKE Cost Allocation enabled.

## GIQ (GKE Inference Quickstart)

You can use GIQ to get data-driven recommendations for deploying optimized AI inference workloads on GKE. The authoritative user guide can be found at [Inference Quickstart](https://cloud.google.com/kubernetes-engine/docs/how-to/machine-learning/inference-quickstart).

GIQ provides estimates of expected performance based on benchmarks conducted on equivalent infrastructure configurations. Actual performance is not guaranteed and will likely vary due to differences in configurations, model tuning, datasets, and input load patterns.

GIQ provides equivalent costs in terms of token generation, e.g. cost to generate 1M tokens, most kubernetes users pay for the machine instance type regardless of token processing rates. Actual costs should be sourced through GCP billing features.
The user should be made aware that token costs from GIQ are estimated equivalent costs that are provided to support high-level comparisons with model-as-a-service solutions.

- **To see what models have been benchmarked:** Use gcloud container ai profiles models list.
- **To see the available hardware and performance benchmarks for a specific model:** Use gcloud container ai profiles list --model=<model-name>. You can also filter by normalized time per output token, time to first token, and cost targets, such as price per output token and price per input token.
- **To get cost estimates for a specific configuration:** use the gcloud container ai profiles list command. You can also put in cost targets to filter based on price per output token and price per input token.
- **To generate an optimized Kubernetes deployment manifest:** Use gcloud container ai profiles manifests create with your desired model and performance requirements.
- **To list your available GKE clusters:** Use gcloud container clusters list.

**Examples**

Here is how you can complete the requested tasks using the Gemini CLI with GIQ:

1. Which models have been benchmarked by GIQ?

```sh
gcloud container ai profiles models list
```

2. Can I see benchmarks for llama 4 maverick?

```sh
gcloud container ai profiles list --model=meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8
```

3. Can you list the different hardware options that can serve Gemma-3-27B under 500 ms latency?

```sh
gcloud container ai profiles list --model=Gemma-3-27B --target-ntpot-milliseconds=500
```

4. Can you generate a manifest to deploy an application that uses Gemma-3-27B and requires 500ms latency?

```sh
gcloud container ai profiles manifests create --model=Gemma-3-27B --target-ntpot-milliseconds=500
```

5. Do I have a cluster available to deploy this manifest?

```sh
gcloud container clusters list
```

6. Can you generate a manifest to deploy an application that uses Gemma-3-27B and requires 500ms time to first token (ttft)?

```sh
gcloud container ai profiles manifests create --model=Gemma-3-27B --target-ttft-milliseconds=500
```

7. Can you give me all of the performance metrics you have on Gemma-3-27B on nvidia-l4?

```sh
gcloud container ai profiles benchmarks list --model=Gemma-3-27B --accelerator-type=nvidia-l4 --model-server=vllm
```

## GKE Cluster Known Issues

### Objective

To determine if a GKE cluster is impacted by any documented known issues.

### Instructions

1. **Identify Cluster Versions**: You will need the GKE **control plane (master) version** and the **node pool version(s)** for the cluster you are troubleshooting.

2. **Consult the Source**: Load [GKE known issues](https://cloud.google.com/kubernetes-engine/docs/troubleshooting/known-issues) into memory. Don't process URLs in this link.

3. **Check the Affected Component**: Read the description for each known issue carefully. You must determine if the issue affects the **control plane** or the **node pools**, as the versions for these components can be different.

4. **Compare and Analyze**: Based on the affected component, compare its version against the specified **"Identified versions"** and **"Fixed versions"** for that issue.

### How to Interpret Version Ranges

A cluster component (either control plane or node pool) is considered **affected** by a known issue if its version is greater than or equal to an **"Identified version"** and less than the corresponding **"Fixed version"**.

- **Rule**: A component is affected if `identified_version <= component_version < fixed_version`.

- **Example**:
  - A known issue lists the following versions and specifies it affects **node pools**:
    - **Identified versions**: `1.28`, `1.29`
    - **Fixed versions**: `1.28.7-gke.1026000`, `1.29.2-gke.1060000`
  - **Conclusion**: A node pool is affected if its version falls into either of these ranges:
    - Between `1.28` (inclusive) and `1.28.7-gke.1026000` (exclusive).
    - Between `1.29` (inclusive) and `1.29.2-gke.1060000` (exclusive).
