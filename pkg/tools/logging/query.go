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

package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"text/template"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/GoogleCloudPlatform/gke-mcp/pkg/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	_ "google.golang.org/genproto/googleapis/cloud/audit" // Import for AuditLog proto so we can convert to JSON.
	"google.golang.org/protobuf/encoding/protojson"
)

// LogQueryRequest defines parameters for querying GCP logs.
type LogQueryRequest struct {
	Query     string    `json:"query" jsonschema:"LQL query string to filter and retrieve log entries. Don't specify time ranges in this filter. Use 'time_range' instead."`
	ProjectID string    `json:"project_id" jsonschema:"GCP project ID to query logs from. Required."`
	TimeRange TimeRange `json:"time_range,omitempty" jsonschema:"Time range for log query. If empty, no restrictions are applied."`
	Since     string    `json:"since,omitempty" jsonschema:"Only return logs newer than a relative duration like 5s, 2m, or 3h. The only supported units are seconds ('s'), minutes ('m'), and hours ('h')."`
	Limit     int       `json:"limit,omitempty" jsonschema:"Maximum number of log entries to return. Cannot be greater than 100. Consider multiple calls if needed. Defaults to 10."`
	Format    string    `json:"format,omitempty" jsonschema:"Go template string to format each log entry. If empty, the full JSON representation is returned. Note that empty fields are not included in the response. Example: '{{.timestamp}} [{{.severity}}] {{.textPayload}}'. It's strongly recommended to use a template to minimize the size of the response and only include the fields you need. Use the get_schema tool before this tool to get information about supported log types and their schemas."`
}

// TimeRange captures an optional start/end window for log queries.
type TimeRange struct {
	StartTime time.Time `json:"start_time" jsonschema:"Start time for log query (RFC3339 format)"`
	EndTime   time.Time `json:"end_time" jsonschema:"End time for log query (RFC3339 format)"`
}

const (
	defaultLimit = 10
	maxLimit     = 100
)

func installQueryLogsTool(s *mcp.Server, conf *config.Config) {
	t := newQueryLogsTool(conf)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "query_logs",
		Description: "Query Google Cloud Platform logs using Logging Query Language (LQL). Before using this tool, it's **strongly** recommended to call the 'get_log_schema' tool to get information about supported log types and their schemas. Logs are returned in ascending order, based on the timestamp (i.e. oldest first).",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, t.queryLogs)
}

type queryLogsTool struct {
	conf *config.Config
}

func newQueryLogsTool(conf *config.Config) *queryLogsTool {
	return &queryLogsTool{
		conf: conf,
	}
}

func (t *queryLogsTool) queryLogs(ctx context.Context, _ *mcp.CallToolRequest, req *LogQueryRequest) (*mcp.CallToolResult, any, error) {
	req.setDefaults()
	if err := req.validate(); err != nil {
		return nil, nil, err
	}
	result, err := t.queryGCPLogs(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: result},
		},
	}, nil, nil
}

func (r *LogQueryRequest) setDefaults() {
	if r.Limit == 0 {
		r.Limit = defaultLimit
	}
}

func (r *LogQueryRequest) validate() error {
	if r.ProjectID == "" {
		return fmt.Errorf("project_id parameter is required")
	}
	if r.Limit > maxLimit {
		return fmt.Errorf("limit parameter cannot be greater than %d", maxLimit)
	}
	if r.Since != "" {
		if _, err := time.ParseDuration(r.Since); err != nil {
			return fmt.Errorf("invalid since parameter: %w", err)
		}
	}
	if (r.TimeRange != TimeRange{}) && r.Since != "" {
		return fmt.Errorf("since parameter cannot be used with time_range")
	}
	if r.Format != "" {
		var err error
		_, err = template.New("log").Parse(r.Format)
		if err != nil {
			return fmt.Errorf("invalid format template: %w", err)
		}
	}
	return nil
}

func (t *queryLogsTool) queryGCPLogs(ctx context.Context, req *LogQueryRequest) (string, error) {
	client, err := logging.NewClient(ctx, option.WithUserAgent(t.conf.UserAgent()))
	if err != nil {
		return "", fmt.Errorf("failed to create logging client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close logging client: %v\n", err)
		}
	}()

	listLogsReq := buildListLogEntriesRequest(req)
	// Request one more than the limit to check for truncation.
	// #nosec G115
	listLogsReq.PageSize = int32(req.Limit + 1)

	resp := client.ListLogEntries(ctx, listLogsReq)

	var entries []*loggingpb.LogEntry
	for {
		entry, err := resp.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to iterate log entries: %v", err)
		}
		entries = append(entries, entry)
		if len(entries) > req.Limit {
			break
		}
	}

	truncated := len(entries) > req.Limit
	if truncated {
		entries = entries[:req.Limit]
	}

	allLogLines := strings.Builder{}
	if len(entries) == 0 {
		allLogLines.WriteString("No log entries found.")
	} else {
		formatter, err := formatterForRequest(req)
		if err != nil {
			return "", fmt.Errorf("failed to create formatter: %w", err)
		}

		for i, entry := range entries {
			if i > 0 {
				allLogLines.WriteString("\n")
			}
			logLine, err := formatter.format(entry)
			if err != nil {
				return "", fmt.Errorf("failed to format log entry: %w", err)
			}
			allLogLines.WriteString(logLine)
		}
	}

	result := fmt.Sprintf("Project ID: %s\nLQL Query:\n```\n%s\n```\nResult:\n\n%s", req.ProjectID, listLogsReq.Filter, allLogLines.String())
	if truncated {
		result += fmt.Sprintf("\n\nWarning: Results truncated. The query returned more than the limit of %d log entries. You can use the `limit` parameter to request more entries (up to %d).", req.Limit, maxLimit)
	}

	return result, nil
}

func buildListLogEntriesRequest(req *LogQueryRequest) *loggingpb.ListLogEntriesRequest {
	filter := req.Query

	if req.Since != "" {
		since, err := time.ParseDuration(req.Since)
		if err != nil {
			return nil
		}
		req.TimeRange = TimeRange{
			StartTime: time.Now().Add(-since),
		}
	}
	// filters based on TimeRange
	{
		var timeFilters []string
		if !req.TimeRange.StartTime.IsZero() {
			timeFilters = append(timeFilters, fmt.Sprintf(`timestamp >= "%s"`, req.TimeRange.StartTime.Format(time.RFC3339)))
		}
		if !req.TimeRange.EndTime.IsZero() {
			timeFilters = append(timeFilters, fmt.Sprintf(`timestamp <= "%s"`, req.TimeRange.EndTime.Format(time.RFC3339)))
		}
		if len(timeFilters) > 0 {
			if filter != "" {
				filter += " AND "
			}
			filter += strings.Join(timeFilters, " AND ")
		}
	}
	return &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", req.ProjectID)},
		Filter:        filter,
		// #nosec G115
		PageSize: int32(req.Limit),
		OrderBy:  "timestamp asc",
	}
}

func formatterForRequest(req *LogQueryRequest) (formatter, error) {
	if req.Format == "" {
		return &jsonFormatter{}, nil
	}

	tmpl, err := template.New("log").Parse(req.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to parse format template: %w", err)
	}
	return &goTemplateFormatter{tmpl: tmpl}, nil
}

type formatter interface {
	format(entry *loggingpb.LogEntry) (string, error)
}

type jsonFormatter struct{}

func (f *jsonFormatter) format(entry *loggingpb.LogEntry) (string, error) {
	m := protojson.MarshalOptions{
		Multiline:       true,
		Indent:          "  ",
		EmitUnpopulated: false,
	}
	logLine, err := m.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("could not marshal log entry to JSON: %w", err)
	}
	return string(logLine), nil
}

type goTemplateFormatter struct {
	tmpl *template.Template
}

func (f *goTemplateFormatter) format(entry *loggingpb.LogEntry) (string, error) {
	b, err := protojson.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("could not marshal log entry to JSON for template: %w", err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return "", fmt.Errorf("could not unmarshal log entry to map for template: %w", err)
	}
	var logLine strings.Builder
	if err := f.tmpl.Execute(&logLine, data); err != nil {
		return "", err
	}
	return logLine.String(), nil
}
