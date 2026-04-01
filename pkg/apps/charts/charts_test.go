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

package charts

import (
	"context"
	"strings"
	"testing"
	"time"

	monitoringpb "cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapTimeseriesDataPoints(t *testing.T) {
	now := time.Now()
	nowUnixMilli := now.UnixMilli()

	testCases := []struct {
		name     string
		input    *monitoringpb.TimeSeriesData
		expected appTimeSeries
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: appTimeSeries{},
		},
		{
			name: "single point with value, timestamp and label",
			input: &monitoringpb.TimeSeriesData{
				LabelValues: []*monitoringpb.LabelValue{
					{Value: &monitoringpb.LabelValue_StringValue{StringValue: "label-1"}},
				},
				PointData: []*monitoringpb.TimeSeriesData_PointData{
					{
						Values: []*monitoringpb.TypedValue{
							{Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 42.5}},
						},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now),
						},
					},
				},
			},
			expected: appTimeSeries{
				Label: "label-1",
				Points: []appTimeSeriesDataPoint{
					{Timestamp: nowUnixMilli, Value: 42.5},
				},
			},
		},
		{
			name: "multiple points in order, multiple labels",
			input: &monitoringpb.TimeSeriesData{
				LabelValues: []*monitoringpb.LabelValue{
					{Value: &monitoringpb.LabelValue_StringValue{StringValue: "foo"}},
					{Value: &monitoringpb.LabelValue_StringValue{StringValue: "bar"}},
				},
				PointData: []*monitoringpb.TimeSeriesData_PointData{
					{
						Values: []*monitoringpb.TypedValue{
							{Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 1.0}},
						},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now),
						},
					},
					{
						Values: []*monitoringpb.TypedValue{
							{Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 2.0}},
						},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now.Add(1 * time.Hour)),
						},
					},
				},
			},
			expected: appTimeSeries{
				Label: "foo bar",
				Points: []appTimeSeriesDataPoint{
					{Timestamp: nowUnixMilli, Value: 1.0},
					{Timestamp: now.Add(1 * time.Hour).UnixMilli(), Value: 2.0},
				},
			},
		},
		{
			name: "data point with no values should be skipped",
			input: &monitoringpb.TimeSeriesData{
				PointData: []*monitoringpb.TimeSeriesData_PointData{
					{
						Values: []*monitoringpb.TypedValue{},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now),
						},
					},
				},
			},
			expected: appTimeSeries{
				Label:  "",
				Points: []appTimeSeriesDataPoint{},
			},
		},
		{
			name: "more than 1 value provided to Values should use the first one",
			input: &monitoringpb.TimeSeriesData{
				PointData: []*monitoringpb.TimeSeriesData_PointData{
					{
						Values: []*monitoringpb.TypedValue{
							{Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 1.23}},
							{Value: &monitoringpb.TypedValue_DoubleValue{DoubleValue: 4.56}},
						},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now),
						},
					},
				},
			},
			expected: appTimeSeries{
				Label: "",
				Points: []appTimeSeriesDataPoint{
					{Timestamp: nowUnixMilli, Value: 1.23},
				},
			},
		},
		{
			name: "value type is int64 should convert to float64",
			input: &monitoringpb.TimeSeriesData{
				PointData: []*monitoringpb.TimeSeriesData_PointData{
					{
						Values: []*monitoringpb.TypedValue{
							{Value: &monitoringpb.TypedValue_Int64Value{Int64Value: 42}},
						},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now),
						},
					},
				},
			},
			expected: appTimeSeries{
				Label: "",
				Points: []appTimeSeriesDataPoint{
					{Timestamp: nowUnixMilli, Value: 42.0},
				},
			},
		},
		{
			name: "value type is different from int64 and double should skip the point",
			input: &monitoringpb.TimeSeriesData{
				PointData: []*monitoringpb.TimeSeriesData_PointData{
					{
						Values: []*monitoringpb.TypedValue{
							{Value: &monitoringpb.TypedValue_StringValue{StringValue: "not-a-number"}},
						},
						TimeInterval: &monitoringpb.TimeInterval{
							EndTime: timestamppb.New(now),
						},
					},
				},
			},
			expected: appTimeSeries{
				Label:  "",
				Points: []appTimeSeriesDataPoint{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := mapTimeseriesDataPoints(tc.input)

			if actual.Label != tc.expected.Label {
				t.Errorf("Label = %q, want %q", actual.Label, tc.expected.Label)
			}

			if len(actual.Points) != len(tc.expected.Points) {
				t.Fatalf("returned %d points, want: %d", len(actual.Points), len(tc.expected.Points))
			}

			for i := range actual.Points {
				if actual.Points[i].Timestamp != tc.expected.Points[i].Timestamp {
					t.Errorf("at index %d: Timestamp = %v, want: %v", i, actual.Points[i].Timestamp, tc.expected.Points[i].Timestamp)
				}
				if actual.Points[i].Value != tc.expected.Points[i].Value {
					t.Errorf("at index %d: Value = %v, want %v", i, actual.Points[i].Value, tc.expected.Points[i].Value)
				}
			}
		})
	}
}

func TestTimeSeriesChartTool_Validation(t *testing.T) {
	h := &handlers{c: &config.Config{}}

	ctx := context.Background()

	testCases := []struct {
		name    string
		args    *timeSeriesChartArgs
		wantErr string
	}{
		{
			name: "empty project id and no default",
			args: &timeSeriesChartArgs{
				Query: "test-query",
				Title: "Test Chart",
			},
			wantErr: "project_id argument cannot be empty",
		},
		{
			name: "empty query",
			args: &timeSeriesChartArgs{
				ProjectID: "test-proj",
				Title:     "Test Chart",
			},
			wantErr: "query argument cannot be empty",
		},
		{
			name: "valid arguments should pass validation and run stub",
			args: &timeSeriesChartArgs{
				ProjectID: "test-proj",
				Query:     "test-query",
				Title:     "Test Chart",
			},
			wantErr: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := h.timeSeriesChart(ctx, nil, tc.args)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("timeSeriesChart() err = nil, want = %q", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Errorf("timeSeriesChart() err = %q, want %q", err.Error(), tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("timeSeriesChart() returned unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("timeSeriesChart() returned nil result")
			}

			textContent := result.Content[0].(*mcp.TextContent)
			if !strings.Contains(textContent.Text, "Rendered time series data") {
				t.Errorf("Unexpected result text: %s", textContent.Text)
			}
		})
	}
}

func TestQueryTimeSeriesTool_Validation(t *testing.T) {
	h := &handlers{c: &config.Config{}}

	ctx := context.Background()

	testCases := []struct {
		name    string
		args    *queryTimeSeriesArgs
		wantErr string
	}{
		{
			name: "empty project id and no default",
			args: &queryTimeSeriesArgs{
				Query: "test-query",
			},
			wantErr: "project_id argument cannot be empty",
		},
		{
			name: "empty query",
			args: &queryTimeSeriesArgs{
				ProjectID: "test-proj",
			},
			wantErr: "query argument cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := h.queryTimeSeries(ctx, nil, tc.args)

			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("queryTimeSeries() err = nil, want = %q", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Errorf("queryTimeSeries() err = %q, want %q", err.Error(), tc.wantErr)
				}
			}
		})
	}
}
