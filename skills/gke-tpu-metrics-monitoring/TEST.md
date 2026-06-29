# Manual Test Plan for GKE TPU Metrics Monitoring Skill

This document describes how to manually verify the GKE TPU Metrics Monitoring skill.

## Test Case 1: Verify PromQL Queries in Cloud Monitoring

1.  Go to the Google Cloud Console.
2.  Navigate to **Monitoring** > **Metrics Explorer**.
3.  Switch to **MQL** or **PromQL** mode (select PromQL).
4.  Try running the following query (replace placeholders):
    ```promql
    kubernetes_io:node_status_condition{monitored_resource="k8s_node", cluster_name="<your-cluster>", condition="Ready"}
    ```
5.  Verify that it returns data if you have GKE system metrics enabled.

## Test Case 2: Verify TPU Duty Cycle Metric

1.  In Metrics Explorer, search for the metric `kubernetes.io/container/accelerator/duty_cycle`.
2.  Ensure it is visible and shows data for your running TPU workloads (which must have `containerPort: 8431` configured).
