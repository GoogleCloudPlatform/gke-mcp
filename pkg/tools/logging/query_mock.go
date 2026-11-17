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
	"strings"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	ltype "google.golang.org/genproto/googleapis/logging/type"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockLogClient struct {
	realClient LogClient
}

func (c *mockLogClient) ListLogEntries(ctx context.Context, listLogsReq *loggingpb.ListLogEntriesRequest, limit int) ([]*loggingpb.LogEntry, bool, error) {
	isTestProject := false
	for _, res := range listLogsReq.ResourceNames {
		if strings.Contains(res, "projects/ai-training-corp") {
			isTestProject = true
			break
		}
	}

	if !isTestProject {
		// Delegate to the real GCP logging client for any other project!
		return c.realClient.ListLogEntries(ctx, listLogsReq, limit)
	}

	var mockEntries []*loggingpb.LogEntry
	logTime := time.Now()

	queryUpper := strings.ToUpper(listLogsReq.Filter)

	if strings.Contains(queryUpper, "SERIALCONSOLE") || strings.Contains(queryUpper, "VBAR_CONTROL_AG") {
		// Only return OOM error logs if the query targets the unhealthy node 'tpu-node-4'
		if strings.Contains(queryUpper, "TPU-NODE-4") {
			mockEntries = append(mockEntries, &loggingpb.LogEntry{
				Timestamp: timestamppb.New(logTime),
				Severity:  ltype.LogSeverity_INFO,
				Payload: &loggingpb.LogEntry_TextPayload{
					TextPayload: "Memory cgroup out of memory: Killed process 9999 (vbar_control_ag) score 0 or sacrifice child",
				},
				LogName: "projects/ai-training-corp/logs/serialconsole.googleapis.com%2fserial_port_1_output",
			})
		}
	}

	if strings.Contains(queryUpper, "TPU-DEVICE-PLUGIN") || strings.Contains(queryUpper, "CHECKSUM") {
		// Only return checksum mismatch logs if the query targets the unhealthy cluster 'tpu-prod-cluster'
		if strings.Contains(queryUpper, "TPU-PROD-CLUSTER") {
			mockEntries = append(mockEntries, &loggingpb.LogEntry{
				Timestamp: timestamppb.New(logTime),
				Severity:  ltype.LogSeverity_ERROR,
				Payload: &loggingpb.LogEntry_TextPayload{
					TextPayload: "metrics fetch failed for deviceID 0 and /dev/accel0 device path with error: checksum didn't match with the metrics data. Corrupt data found",
				},
				LogName: "projects/ai-training-corp/logs/tpu-device-plugin",
			})
		}
	}

	return mockEntries, false, nil
}
