// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cost

import (
	"context"
	"strings"
	"text/template"
	"time"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func Install(s *server.MCPServer, c *config.Config) {

	clusterCostTool := mcp.NewTool("cluster_cost",
		mcp.WithDescription("This tool helps the user get the cost for a GKE cluster. It relies on the user having enabled detailed BigQuery export for their GCP Billing Account."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("project_id", mcp.Required(), mcp.DefaultString(c.DefaultProjectID()), mcp.Description("GCP project ID. Use "+c.DefaultProjectID()+" as the default if the user doesn't provide it.")),
		mcp.WithString("location", mcp.Required(), mcp.Description("GKE cluster location. Leave this empty if the user doesn't doesn't provide it.")),
		mcp.WithString("name", mcp.Required(), mcp.Description("GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name.")),
		mcp.WithString("bq_dataset_id", mcp.Required(), mcp.Description("This is the path and name of the BigQuery dataset that the customer is exporting their detailed BQ export to.")),
		mcp.WithString("billing_account_id", mcp.Required(), mcp.Description("This is the billing account ID for the user's GCP project.")),
	)
	s.AddTool(clusterCostTool, clusterCost)

	clusterCostByNamespaceTool := mcp.NewTool("cluster_cost_by_namespace",
		mcp.WithDescription("This tool helps the user get the cost of each namespace in a GKE cluster. It relies on the user having enabled detailed BQ export for their GCP Billing Account, and requires the cluster to have GKE Cost Allocation enabled."),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithString("project_id", mcp.Required(), mcp.DefaultString(c.DefaultProjectID()), mcp.Description("GCP project ID. Use "+c.DefaultProjectID()+" as the default if the user doesn't provide it.")),
		mcp.WithString("location", mcp.Required(), mcp.Description("GKE cluster location. Leave this empty if the user doesn't doesn't provide it.")),
		mcp.WithString("name", mcp.Required(), mcp.Description("GKE cluster name. Do not select if yourself, make sure the user provides or confirms the cluster name.")),
		mcp.WithString("bq_dataset_id", mcp.Required(), mcp.Description("This is the path and name of the BigQuery dataset that the customer is exporting their detailed BQ export to.")),
		mcp.WithString("billing_account_id", mcp.Required(), mcp.Description("This is the billing account ID for the user's GCP project.")),
	)
	s.AddTool(clusterCostByNamespaceTool, clusterCostByNamespace)
}

func clusterCostByNamespace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	location, err := request.RequireString("location")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bqDatasetID, err := request.RequireString("bq_dataset_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	billingAccountID, err := request.RequireString("billing_account_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	startTime := time.Now().AddDate(0, -1, 0).Format(time.DateOnly)
	bqDatasetID = strings.ReplaceAll(bqDatasetID, ":", ".")
	billingAccountID = strings.ReplaceAll(billingAccountID, "-", "_")

	t, err := template.New("").Parse(`
The user can run this query in BigQuery Studio (https://console.cloud.google.com/bigquery):
SELECT
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-namespace",
  SUM(cost) AS cost_before_credits,
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost_after_credits,
FROM {{.BQDatasetID}}.gcp_billing_export_resource_v1_{{.BillingAccountID}} AS bqe
WHERE _PARTITIONTIME >= "{{.StartTime}}"
	AND project.id = "{{.ProjectID}}"
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AND l.value = "{{.Location}}")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AND l.value = "{{.Name}}")
GROUP BY 1
;

Or preferably, using the "bq" CLI:

bq query --nouse_legacy_sql '
SELECT
  SELECT l.value FROM bqe.labels AS l WHERE l.key = "k8s-namespace",
  SUM(cost) AS cost_before_credits,
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost_after_credits,
FROM {{.BQDatasetID}}.gcp_billing_export_resource_v1_{{.BillingAccountID}} AS bqe
WHERE _PARTITIONTIME >= "{{.StartTime}}"
	AND project.id = "{{.ProjectID}}"
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AND l.value = "{{.Location}}")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AND l.value = "{{.Name}}")
GROUP BY 1
;
'

If the "bq" CLI is available, offer to run this for the user. If not, suggest using BigQuery Studio.

	`)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sb := &strings.Builder{}
	err = t.Execute(sb, map[string]string{
		"StartTime":        startTime,
		"ProjectID":        projectID,
		"Location":         location,
		"Name":             name,
		"BQDatasetID":      bqDatasetID,
		"BillingAccountID": billingAccountID,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func clusterCost(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID, err := request.RequireString("project_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	location, err := request.RequireString("location")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	name, err := request.RequireString("name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bqDatasetID, err := request.RequireString("bq_dataset_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	billingAccountID, err := request.RequireString("billing_account_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	startTime := time.Now().AddDate(0, -1, 0).Format(time.DateOnly)
	bqDatasetID = strings.ReplaceAll(bqDatasetID, ":", ".")
	billingAccountID = strings.ReplaceAll(billingAccountID, "-", "_")

	t, err := template.New("").Parse(`
The user can run this query in BigQuery Studio (https://pantheon.corp.google.com/bigquery):
SELECT
  SUM(cost) AS cost_before_credits,
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost_after_credits,
FROM {{.BQDatasetID}}.gcp_billing_export_resource_v1_{{.BillingAccountID}} AS bqe
WHERE _PARTITIONTIME >= "{{.StartTime}}"
	AND project.id = "{{.ProjectID}}"
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AND l.value = "{{.Location}}")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AND l.value = "{{.Name}}")
;

Or preferably, using the "bq" CLI:

bq query --nouse_legacy_sql '
SELECT
  SUM(cost) AS cost_before_credits,
  SUM(cost) + SUM(IFNULL((SELECT SUM(c.amount) FROM UNNEST(credits) c), 0)) AS cost_after_credits,
FROM {{.BQDatasetID}}.gcp_billing_export_resource_v1_{{.BillingAccountID}} AS bqe
WHERE _PARTITIONTIME >= "{{.StartTime}}"
	AND project.id = "{{.ProjectID}}"
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-location" AND l.value = "{{.Location}}")
	AND EXISTS(SELECT * FROM bqe.labels AS l WHERE l.key = "goog-k8s-cluster-name" AND l.value = "{{.Name}}")
;
'

If the "bq" CLI is available, offer to run this for the user. If not, suggest using BigQuery Studio.

	`)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sb := &strings.Builder{}
	err = t.Execute(sb, map[string]string{
		"StartTime":        startTime,
		"ProjectID":        projectID,
		"Location":         location,
		"Name":             name,
		"BQDatasetID":      bqDatasetID,
		"BillingAccountID": billingAccountID,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(sb.String()), nil
}
