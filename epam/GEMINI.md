# GKE MCP Extension for Gemini CLI

This document provides instructions for an AI agent on how to use the available tools to manage Google Kubernetes Engine (GKE) resources.

## Guiding Principles

*   **Prefer Native Tools:** Always prefer to use the tools provided by this extension (e.g., `list_clusters`, `get_cluster`) instead of shelling out to `gcloud` or `kubectl` for the same functionality. This ensures better-structured data and more reliable execution.
*   **Clarify Ambiguity:** Do not guess or assume values for required parameters like cluster names or locations. If the user's request is ambiguous, ask clarifying questions to confirm the exact resource they intend to interact with.
*   **Use Defaults:** If a `project_id` is not specified by the user, you can use the default value configured in the environment.
*   **Verify Commands:** Before providing any command to the user， verify it is correct and appropriate for the user's request. You can search online or refer to https://cloud.google.com/sdk/gcloud for gcloud documentations.
*   **Verbosity:** In the end of response add related links which were used to form a response. 
*   **Table Investigation:** If in document search Table appears - read it in JSON format to correctly interpred provided data.

## Authentication

Some MCP tools required [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). If they return an "Unauthenticated" error, tell the user to run `gcloud auth application-default login` and try again. This is an interactive command and must be run manually outside the AI.

## GKE Logs

*   When searching for GKE logs, always use the `query_logs` tool to fetch them. It's also **strongly** recommended to call the `get_log_schema` tool before building or running a query to obtain information about the log schema, as well as sample queries. This information is useful when building Cloud Logging LQL queries.

*   When using time ranges, make sure you check the current time and date if the range is relative to the current time or date.

*   When searching log entries for a single cluster, **always** include an LQL filter clause for the project ID, cluster name, and cluster location. Note that filtering by project ID is needed even if the project ID is specified in the `query_logs` request, as depending on the log ingention configuration, multiple logs with same name and location can be ingested into the same project.

*   If you need help understanding LQL syntax, consider fetching https://cloud.google.com/logging/docs/view/logging-query-language to learn more about it.

## GKE Monitoring

When users ask a question about the Monitoring or monitored resource types, the following instructions could be applied:

*  Please use the tool `list_monitored_resource_descriptors` to get all monitored resource descriptors
*  After getting all the monitored resource, if the user ask for GKE specific ones, please filter the output and only include the GKE related ones
   ** Full GKE related monitored resources are the one contains `gke` or `k8s` or `container.googleapis.com`

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

## GIQ (GKE Inference Quickstart)

You can use GIQ to get data-driven recommendations for deploying optimized AI inference workloads on GKE. The authoritative user guide can be found at https://cloud.google.com/kubernetes-engine/docs/how-to/machine-learning/inference-quickstart. 

GIQ provides estimates of expected performance based on benchmarks conducted on equivalent infrastructure configurations. Actual performance is not guaranteed and will likely vary due to differences in configurations, model tuning, datasets, and input load patterns.  

GIQ provides equivalent costs in terms of token generation, e.g. cost to generate 1M tokens, most kubernetes users pay for the machine instance type regardless of token processing rates. Actual costs should be sourced through GCP billing features. The user should be made aware that token costs from GIQ are estimated equivalent costs that are provided to support high-level comparisons with model-as-a-service solutions.

* **To see what models have been benchmarked:** Use gcloud alpha container ai profiles models list.
* **To see the available hardware and performance benchmarks for a specific model:** Use gcloud alpha container ai profiles accelerators list --model=<model-name>. You can also filter by latency.
* **To get cost estimates for a specific configuration:** Add the --cost-usage-type flag to the gcloud alpha container ai profiles accelerators list command.
* **To generate an optimized Kubernetes deployment manifest:** Use gcloud alpha container ai profiles manifests create with your desired model and performance requirements.
* **To list your available GKE clusters:** Use gcloud container clusters list.

**Examples**

Here is how you can complete the requested tasks using the Gemini CLI with GIQ:

1. Which models have been benchmarked by GIQ?
```sh
gcloud alpha container ai profiles models list
```

2. Can I see benchmarks for llama 4 maverick?
```sh
gcloud alpha container ai profiles accelerators list --model=meta-llama/Llama-4-Maverick-17B-128E-Instruct-FP8
```

3. Can you list the different hardware options that can serve Gemma-3-27B under 500 ms latency?
```sh
gcloud alpha container ai profiles accelerators list --model=Gemma-3-27B --target-ntpot-milliseconds=500
```

4. Can you generate a manifest to deploy an application that uses Gemma-3-27B and requires 500ms latency?
```sh
gcloud alpha container ai profiles manifests create --model=Gemma-3-27B --target-ntpot-milliseconds=500 --output-file=gemma_deployment.yaml
```

5. Do I have a cluster available to deploy this manifest?
```sh
gcloud container clusters list
```

## GKE Cluster Known Issues

### Objective
To determine if a GKE cluster is impacted by any documented known issues.

### Instructions

1.  **Identify Cluster Versions**: You will need the GKE **control plane (master) version** and the **node pool version(s)** for the cluster you are troubleshooting.

2.  **Consult the Source**: Load https://cloud.google.com/kubernetes-engine/docs/troubleshooting/known-issues into memory. Don't process URLs in this link.

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

## GCP Storage Pools

When a user asks about Storage Pools you have to check an official documentation at first.

- Disk Types - [Disk types](https://cloud.google.com/compute/docs/disks).
- Storage Pool usage - [Storage Pool usage](https://cloud.google.com/compute/docs/disks/storage-pools).
- Create Storage Pool - [Create Storage Pool](https://cloud.google.com/compute/docs/disks/create-storage-pools).
- Manage Storage Pool - [](https://cloud.google.com/compute/docs/disks/manage-storage-pools).

## GKE Storage and Hyperdisks

When a user asks about Hyperdisk or storage operations, you must first consult the official documentation from the list below to provide accurate and up-to-date information.
Official Documentation:

- How to create Hyperdisk - [Create Hyperdisk](https://cloud.google.com/compute/docs/disks/add-hyperdisk).
- Hyperdisk ML storage class - [Hyperdisk ML storage class](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/hyperdisk-ml).
- Hyperdisk Balanced, Throughput, Extreme or Balanced High Availability storage class - [Hyperdisk storage class](https://cloud.googlegoogle.com/kubernetes-engine/docs/how-to/persistent-volumes/hyperdisk).
- Hyperdisk overview and available disk types across Compute Engine instances - [Hyperdisk overview](https://cloud.google.com/compute/docs/disks/hyperdisks).
- Persistent disk overview and available disk types - [Persistent disk overview](https://cloud.google.com/compute/docs/disks/persistent-disks).
- Regional persistent disks storage class - [Regional persistent disks](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/regional-pd).
- Hyperdisk Performance limitations - [Hyperdisk Performance limitations](https://cloud.google.com/compute/docs/disks/hyperdisk-perf-limits#limits-by-hd-type).
- Hyperdisk ML storage class - [Hyperdisk ML storage class](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/hyperdisk-ml).
- Migration to Hyperdisks - [Hyperdisk migration](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/hyperdisk).
- How to use existing disks in GKE - [Pre-existing PD](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/preexisting-pd).
- How to create disk snapshots - [Disk snapshots creation](https://cloud.google.com/compute/docs/disks/create-snapshots).
- Persistent disks and hyperdisks pricing - [Disk and image pricing](https://cloud.google.com/compute/disks-image-pricing).

If you cannot find the required information in these documents, you may use Google Search and follow external links if necessary to complete the user request accurately.

If the user requests any action or change involving disks, such as creating, deleting, attaching, detaching, migrating, resizing, or modifying Hyperdisks or persistent disks:
* Do not perform any changes immediately.
* First, present a concise, clearly structured plan that includes:
    - The goals of the action.
    - Short explanation of each step.
    - A summary of steps you intend to take.
    - Potential risks or consequences (e.g., service downtime, data loss).
    - Any prerequisites that must be satisfied.

Never execute actions against production infrastructure without explicit confirmation. The system operates in production, and data loss is unacceptable.
Prefer using gcloud CLI commands for all actions.
Use disk snapshots for backups or as part of any disk migration process.
Always plan before acting. Always confirm before executing.

## Storage options
When a user asks about the storage types to use for his particular workload. Use information from included links for  more detailed response. Next check if instance support this type of storage. Provide a response naming a machine type taken from user envirinment and highlight storage option in use then provide a list of acceptable options   

- Storage types - [Storage types](https://cloud.google.com/blog/products/storage-data-transfer/pick-the-right-storage-option-on-google-cloud).
- Machines family comparison - [Machines family comparison](https://cloud.google.com/compute/docs/machine-resource).

Use Google search if you can't find required information in these web pages.

## GS Fuse
  When user asks about Object Storage connection options request the name of storage bucket and a cluster you want to connect to, or provide a general info. 
  When creating PersistantVolume manifest use folloving values for "mountOptions": 
  [mountOptions: "implicit-dirs,metadata-cache:ttl-secs:-1,file-cache:max-size-mb:-1,file-cache:cache-file-for-range-read:true,file-cache:enable-parallel-downloads:true"]

- GS FUSE driver - [GS FUSE driver](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/cloud-storage-fuse-csi-driver).
- About - [About](https://cloud.google.com/kubernetes-engine/docs/concepts/cloud-storage-fuse-csi-driver).
- Setup - [Setup](https://cloud.google.com/kubernetes-engine/docs/how-to/cloud-storage-fuse-csi-driver-setup).
- Mount - [Mount CSI driver](https://cloud.google.com/kubernetes-engine/docs/how-to/cloud-storage-fuse-csi-driver-ephemeral, https://cloud.google.com/kubernetes-engine/docs/how-to/cloud-storage-fuse-csi-driver-pv).
## Terraform Cluster creation 
