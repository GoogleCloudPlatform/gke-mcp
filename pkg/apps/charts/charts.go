// Copyright 2026 Google LLC
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

// Package charts provides an MCP app for rendering Google Cloud Monitoring charts.
package charts

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	htmlFilePath = "ui/dist/apps/timeserieschart/index.html"
	resourceURI  = "ui://monitoring_time_series_chart/index.html"
	mimeType     = "text/html;profile=mcp-app"
)

type handlers struct {
	c *config.Config
}

type timeSeriesChartArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Query     string `json:"query" jsonschema:"Required. The query in the Monitoring Query Language (MQL) format."`
	Title     string `json:"title,omitempty" jsonschema:"Optional. The title to display for the time series chart."`
	XLegend   string `json:"x_legend,omitempty" jsonschema:"Optional. The legend/label for the X-axis (e.g., 'Time', 'Date')."`
	YLegend   string `json:"y_legend,omitempty" jsonschema:"Optional. The legend/label for the Y-axis (e.g., 'CPU Usage (%)', 'Memory (GiB)')."`
}

type queryTimeSeriesArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Query     string `json:"query" jsonschema:"Required. The query in the Monitoring Query Language (MQL) format."`
}

type queryTimeSeriesResponse struct {
	Data []appTimeSeries `json:"data"`
}

type mqlValidatorArgs struct {
	ProjectID string `json:"project_id,omitempty" jsonschema:"GCP project ID. Use the default if the user doesn't provide it."`
	Query     string `json:"query" jsonschema:"Required. The test query in the MQL format to validate."`
}

type validationResult struct {
	Status       string `json:"status"`
	Query        string `json:"query"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type appTimeSeries struct {
	Label  string                   `json:"label,omitempty"`
	Points []appTimeSeriesDataPoint `json:"points,omitempty"`
}

type appTimeSeriesDataPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// Install registers monitoring tools that require 'apps' extension support.
func Install(_ context.Context, s *mcp.Server, c *config.Config) error {
	h := &handlers{
		c: c,
	}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "monitoring_time_series_chart",
		Description: "Interactive tool to display time series data using a React Chart. ALWAYS favor using this tool to query metrics rather than outputting raw values so the user gets a visualization. MUST Call `mql_validator` FIRST to catch syntax issues or metric anomalies before running this tool.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "GCP project ID. Use the default if the user doesn't provide it.",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Required. The query in the Monitoring Query Language (MQL) format. Explicitly append `| within 1h` or similar if you intend to fetch historical points, otherwise it will default to 1h. Ensure you use MQL tools to convert raw metrics to human-readable formats like percentages or where applicable.",
				},
				"title": map[string]interface{}{
					"type":        "string",
					"description": "Optional. The title to display for the time series chart.",
				},
				"x_legend": map[string]interface{}{
					"type":        "string",
					"description": "Optional. The legend/label for the X-axis (e.g., 'Time', 'Date').",
				},
				"y_legend": map[string]interface{}{
					"type":        "string",
					"description": "Optional. The legend/label for the Y-axis (e.g., 'CPU Usage (%)', 'Memory (GiB)').",
				},
			},
			"required": []string{"query"},
		},
		Meta: mcp.Meta{
			"ui": map[string]interface{}{
				"resourceUri": resourceURI,
			},
		},
	}, h.timeSeriesChart)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "query_time_series",
		Description: "Internal app tool. Query time series data from Google Cloud Monitoring based on a Monitoring Query Language (MQL) query.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "GCP project ID. Use the default if the user doesn't provide it.",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Required. The query in the Monitoring Query Language (MQL) format. Explicitly append `| within 1h` or similar if you intend to fetch historical points, otherwise it will default to 1h. Ensure you use MQL scale operations (e.g., `| scale 'GiB'`, `| mul 100`) to convert raw metrics to human-readable formats like percentages or GiB where applicable.",
				},
			},
			"required": []string{"query"},
		},
		Meta: mcp.Meta{
			"ui": map[string]interface{}{
				"visibility": []string{"app"},
			},
		},
	}, h.queryTimeSeries)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "mql_validator",
		Description: "A helper tool to validate Monitoring Query Language (MQL) metric strings. MUST be called immediately before calling `monitoring_time_series_chart` or `query_time_series` to ensure the MQL statement compiles correctly. It fetches 1 page of data to verify syntactical and logical correctness. Returns the original string on success, or an error payload explaining the misconfiguration.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"project_id": map[string]interface{}{
					"type":        "string",
					"description": "GCP project ID. Use the default if the user doesn't provide it.",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Required. The query in the MQL format to validate.",
				},
			},
			"required": []string{"query"},
		},
		Meta: mcp.Meta{
			"ui": map[string]interface{}{
				"visibility": []string{"model"},
			},
		},
	}, h.mqlValidator)

	s.AddResource(&mcp.Resource{
		Name:     "Time Series Chart UI",
		URI:      resourceURI,
		MIMEType: mimeType,
	}, func(ctx context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		htmlContent, err := os.ReadFile(htmlFilePath)

		if err != nil {
			return nil, fmt.Errorf("failed to read UI file at %s: %w", htmlFilePath, err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      resourceURI,
					MIMEType: mimeType,
					Text:     string(htmlContent),
				},
			},
		}, nil
	})

	return nil
}

func (h *handlers) timeSeriesChart(ctx context.Context, _ *mcp.CallToolRequest, args *timeSeriesChartArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.ProjectID == "" {
		return nil, nil, fmt.Errorf("project_id argument cannot be empty")
	}
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query argument cannot be empty")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Rendered time series data in UI component."},
		},
	}, nil, nil
}

func (h *handlers) queryTimeSeries(ctx context.Context, _ *mcp.CallToolRequest, args *queryTimeSeriesArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.ProjectID == "" {
		return nil, nil, fmt.Errorf("project_id argument cannot be empty")
	}
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query argument cannot be empty")
	}

	data, err := queryMonitoringData(ctx, h.c, args.ProjectID, args.Query)
	if err != nil {
		return nil, nil, err
	}

	var series []appTimeSeries
	for _, resp := range data {
		series = append(series, mapTimeseriesDataPoints(resp))
	}

	return &mcp.CallToolResult{
		StructuredContent: queryTimeSeriesResponse{
			Data: series,
		},
	}, nil, nil
}

func mapTimeseriesDataPoints(resp *monitoringpb.TimeSeriesData) appTimeSeries {
	var labelParts []string
	for _, lv := range resp.GetLabelValues() {
		if v := lv.GetStringValue(); v != "" {
			labelParts = append(labelParts, v)
		}
	}
	label := strings.Join(labelParts, " ")

	var pts []appTimeSeriesDataPoint
	for _, p := range resp.GetPointData() {
		if len(p.GetValues()) == 0 {
			continue // if values is not presented, should not be formed
		}
		var val float64
		switch v := p.GetValues()[0].GetValue().(type) {
		case *monitoringpb.TypedValue_DoubleValue:
			val = v.DoubleValue
		case *monitoringpb.TypedValue_Int64Value:
			val = float64(v.Int64Value)
		default:
			continue // skip if value type is not supported by the time series chart
		}

		pts = append(pts, appTimeSeriesDataPoint{
			Timestamp: p.GetTimeInterval().GetEndTime().AsTime().UnixMilli(),
			Value:     val,
		})
	}
	return appTimeSeries{
		Label:  label,
		Points: pts,
	}
}

func (h *handlers) mqlValidator(ctx context.Context, _ *mcp.CallToolRequest, args *mqlValidatorArgs) (*mcp.CallToolResult, any, error) {
	if args.ProjectID == "" {
		args.ProjectID = h.c.DefaultProjectID()
	}
	if args.ProjectID == "" {
		return nil, nil, fmt.Errorf("project_id argument cannot be empty")
	}
	if args.Query == "" {
		return nil, nil, fmt.Errorf("query argument cannot be empty")
	}

	c, err := monitoring.NewQueryClient(ctx, option.WithUserAgent(h.c.UserAgent()), option.WithQuotaProject(args.ProjectID))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create monitoring query client: %w", err)
	}
	defer func() {
		if closeErr := c.Close(); closeErr != nil {
			log.Printf("Failed to close monitoring query client: %v\n", closeErr)
		}
	}()

	req := &monitoringpb.QueryTimeSeriesRequest{
		Name:  fmt.Sprintf("projects/%s", args.ProjectID),
		Query: args.Query,
	}

	it := c.QueryTimeSeries(ctx, req)

	// Fetch the first page just to validate execution.
	_, _, err = it.InternalFetch(1, "")

	var result validationResult
	var isError bool

	if err != nil {
		result = validationResult{
			Status:       "INVALID",
			Query:        args.Query,
			ErrorMessage: fmt.Sprintf("MQL validation failed:\n%v", err),
		}
		isError = true
	} else {
		// Succesful compilation and execution
		result = validationResult{
			Status: "VALID",
			Query:  args.Query,
		}
	}

	return &mcp.CallToolResult{
		IsError:           isError,
		StructuredContent: result,
	}, nil, nil
}

func queryMonitoringData(ctx context.Context, cfg *config.Config, projectID, query string) ([]*monitoringpb.TimeSeriesData, error) {
	c, err := monitoring.NewQueryClient(ctx, option.WithUserAgent(cfg.UserAgent()), option.WithQuotaProject(projectID))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Printf("Failed to close monitoring query client: %v\n", err)
		}
	}()

	req := &monitoringpb.QueryTimeSeriesRequest{
		Name:  fmt.Sprintf("projects/%s", projectID),
		Query: query,
	}

	it := c.QueryTimeSeries(ctx, req)
	var data []*monitoringpb.TimeSeriesData
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, resp)
	}
	return data, nil
}
