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

// Package trace provides MCP tools for querying Google Cloud Trace.
package trace

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/cloudtrace/v1"
	"google.golang.org/api/option"

	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Install registers trace tools with the MCP server.
func Install(_ context.Context, s *mcp.Server, c *config.Config) error {
	h := &handlers{c: c}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "query_traces",
		Description: "Query Google Cloud Trace to retrieve traces for troubleshooting latency or distributed requests. You can specify time ranges, limits, and a filter.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, h.queryTraces)
	return nil
}

type handlers struct {
	c   *config.Config
	svc *cloudtrace.Service
	mu  sync.Mutex
}

func (h *handlers) getService(ctx context.Context) (*cloudtrace.Service, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.svc != nil {
		return h.svc, nil
	}
	svc, err := cloudtrace.NewService(ctx, option.WithUserAgent(h.c.UserAgent()))
	if err != nil {
		return nil, fmt.Errorf("failed to create cloudtrace service: %w", err)
	}
	h.svc = svc
	return h.svc, nil
}

// TimeRange captures an optional start/end window for trace queries.
type TimeRange struct {
	StartTime time.Time `json:"start_time" jsonschema:"Start time for trace query (RFC3339 format)"`
	EndTime   time.Time `json:"end_time" jsonschema:"End time for trace query (RFC3339 format)"`
}

type queryTracesArgs struct {
	ProjectID string    `json:"project_id" jsonschema:"Required. GCP project ID to query traces from."`
	TimeRange TimeRange `json:"time_range,omitempty" jsonschema:"Time range for trace query. If empty, uses 'since'. Cannot be used simultaneously with 'since'."`
	Since     string    `json:"since,omitempty" jsonschema:"Only return traces newer than a relative duration like 5s, 2m, or 3h. Default: '1h' if not specified and time_range is not provided."`
	Limit     int64     `json:"limit,omitempty" jsonschema:"Maximum number of traces to return. Cannot be greater than 100. Defaults to 10."`
	Filter    string    `json:"filter,omitempty" jsonschema:"An optional filter against labels for the request. For example: '+root:http' or '+span:NAME'."`
	MaxSpans  int       `json:"max_spans,omitempty" jsonschema:"Maximum number of spans to return per trace to avoid context limits. Defaults to 5."`
}

func (r *queryTracesArgs) setDefaults(_ *config.Config) {
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.MaxSpans <= 0 {
		r.MaxSpans = 5
	}
	if r.Since == "" && r.TimeRange.StartTime.IsZero() && r.TimeRange.EndTime.IsZero() {
		r.Since = "1h"
	}
}

func (r *queryTracesArgs) validate() error {
	if r.ProjectID == "" {
		return fmt.Errorf("project_id argument is required")
	}
	if r.Limit > 100 {
		return fmt.Errorf("limit cannot be greater than 100")
	}
	if r.MaxSpans > 100 {
		return fmt.Errorf("max_spans cannot be greater than 100")
	}
	if r.Since != "" {
		d, err := time.ParseDuration(r.Since)
		if err != nil {
			return fmt.Errorf("invalid since duration: %w", err)
		}
		if d <= 0 {
			return fmt.Errorf("since duration must be positive")
		}
	}
	if (!r.TimeRange.StartTime.IsZero() || !r.TimeRange.EndTime.IsZero()) && r.Since != "" {
		return fmt.Errorf("since parameter cannot be used with time_range")
	}
	if !r.TimeRange.StartTime.IsZero() || !r.TimeRange.EndTime.IsZero() {
		if r.TimeRange.StartTime.IsZero() || r.TimeRange.EndTime.IsZero() {
			return fmt.Errorf("both start_time and end_time must be provided when using time_range")
		}
		if !r.TimeRange.StartTime.Before(r.TimeRange.EndTime) {
			return fmt.Errorf("start_time must precede end_time")
		}
	}
	return nil
}

func (h *handlers) queryTraces(ctx context.Context, _ *mcp.CallToolRequest, args *queryTracesArgs) (*mcp.CallToolResult, any, error) {
	args.setDefaults(h.c)
	if err := args.validate(); err != nil {
		return nil, nil, err
	}

	svc, err := h.getService(ctx)
	if err != nil {
		return nil, nil, err
	}

	req := svc.Projects.Traces.List(args.ProjectID)
	req.PageSize(args.Limit)

	// By default, the API uses a "MINIMAL" view that omits all spans to save bandwidth.
	// We force "COMPLETE" so the payload includes the span tree, latencies, and labels.
	// Documented at: https://cloud.google.com/trace/docs/reference/v1/rest/v1/projects.traces/list#viewtype
	req.View("COMPLETE")

	if args.Filter != "" {
		req.Filter(args.Filter)
	}

	if args.Since != "" {
		d, err := time.ParseDuration(args.Since)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid since duration: %v", err)
		}
		now := time.Now().UTC()
		startTime := now.Add(-d)
		req.StartTime(startTime.Format(time.RFC3339Nano))
		req.EndTime(now.Format(time.RFC3339Nano))
	} else if !args.TimeRange.StartTime.IsZero() || !args.TimeRange.EndTime.IsZero() {
		req.StartTime(args.TimeRange.StartTime.UTC().Format(time.RFC3339Nano))
		req.EndTime(args.TimeRange.EndTime.UTC().Format(time.RFC3339Nano))
	}

	resp, err := req.Context(ctx).Do()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query traces: %v", err)
	}

	if resp == nil || len(resp.Traces) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("No traces found in project %s.\n", args.ProjectID)},
			},
		}, nil, nil
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Found %d traces in project %s:\n\n", len(resp.Traces), args.ProjectID)

	var printedCount int64
	for _, t := range resp.Traces {
		if t == nil {
			continue
		}
		fmt.Fprintf(&sb, "=== Trace ID: %s ===\n", t.TraceId)

		var spanCount int
		if t.Spans != nil {
			sort.Slice(t.Spans, func(i, j int) bool {
				if t.Spans[i] == nil || t.Spans[j] == nil {
					return t.Spans[i] != nil
				}
				t1, err1 := time.Parse(time.RFC3339Nano, t.Spans[i].StartTime)
				t2, err2 := time.Parse(time.RFC3339Nano, t.Spans[j].StartTime)
				if err1 != nil || err2 != nil {
					return t.Spans[i].StartTime < t.Spans[j].StartTime
				}
				return t1.Before(t2)
			})
			for _, span := range t.Spans {
				if span == nil {
					continue
				}
				spanCount++
				if spanCount > args.MaxSpans {
					// limit spans per trace to avoid blowing up context window
					sb.WriteString("  ... (additional spans truncated)\n")
					break
				}
				fmt.Fprintf(&sb, "  - Span: %s (ID: %016x)\n", span.Name, span.SpanId)

				start, errStart := time.Parse(time.RFC3339Nano, span.StartTime)
				end, errEnd := time.Parse(time.RFC3339Nano, span.EndTime)
				if errStart == nil && errEnd == nil {
					fmt.Fprintf(&sb, "    Duration: %v (%s -> %s)\n", end.Sub(start), span.StartTime, span.EndTime)
				} else {
					fmt.Fprintf(&sb, "    Duration: %s -> %s\n", span.StartTime, span.EndTime)
				}

				if len(span.Labels) > 0 {
					sb.WriteString("    Labels:\n")
					keys := make([]string, 0, len(span.Labels))
					for k := range span.Labels {
						keys = append(keys, k)
					}
					sort.Strings(keys)
					for _, k := range keys {
						fmt.Fprintf(&sb, "      %s: %s\n", k, span.Labels[k])
					}
				}
			}
		}
		sb.WriteString("\n")

		printedCount++
		if printedCount >= args.Limit {
			break
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: sb.String()},
		},
	}, nil, nil
}
