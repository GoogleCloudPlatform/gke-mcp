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

package monitoring

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type metricInfo struct {
	Kind      string
	ValueType string
}

var metricCatalog = map[string]metricInfo{
	"autoscaler_container_cpu_per_replica_recommended_request_cores":  {Kind: "GAUGE", ValueType: "DOUBLE"},
	"autoscaler_container_memory_per_replica_recommended_request_bytes": {Kind: "GAUGE", ValueType: "INT64"},
	"autoscaler_latencies_per_hpa_recommendation_scale_latency_seconds": {Kind: "GAUGE", ValueType: "DOUBLE"},
	"container_cpu_core_usage_time":                                    {Kind: "CUMULATIVE", ValueType: "DOUBLE"},
	"container_cpu_request_utilization":                                {Kind: "GAUGE", ValueType: "DOUBLE"},
	"container_ephemeral_storage_used_bytes":                           {Kind: "GAUGE", ValueType: "INT64"},
	"container_memory_limit_utilization":                               {Kind: "GAUGE", ValueType: "DOUBLE"},
	"container_memory_used_bytes":                                      {Kind: "GAUGE", ValueType: "INT64"},
	"container_restart_count":                                          {Kind: "CUMULATIVE", ValueType: "INT64"},
	"jobset_proxy_runtime_goodput":                                     {Kind: "GAUGE", ValueType: "DOUBLE"},
	"jobset_scheduling_goodput":                                        {Kind: "GAUGE", ValueType: "DOUBLE"},
	"jobset_times_between_interruptions":                               {Kind: "GAUGE", ValueType: "DISTRIBUTION"},
	"jobset_times_to_recover":                                          {Kind: "GAUGE", ValueType: "DISTRIBUTION"},
	"jobset_uptime":                                                    {Kind: "GAUGE", ValueType: "DOUBLE"},
	"node_cpu_allocatable_utilization":                                 {Kind: "GAUGE", ValueType: "DOUBLE"},
	"node_cpu_core_usage_time":                                         {Kind: "CUMULATIVE", ValueType: "DOUBLE"},
	"node_ephemeral_storage_used_bytes":                                {Kind: "GAUGE", ValueType: "INT64"},
	"node_interruption_count":                                          {Kind: "GAUGE", ValueType: "INT64"},
	"node_memory_used_bytes":                                           {Kind: "GAUGE", ValueType: "INT64"},
	"node_network_received_bytes_count":                                {Kind: "CUMULATIVE", ValueType: "INT64"},
	"node_pool_accelerator_times_to_recover":                           {Kind: "GAUGE", ValueType: "DISTRIBUTION"},
	"node_pool_interruption_count":                                     {Kind: "GAUGE", ValueType: "INT64"},
	"node_pool_multi_host_available":                                   {Kind: "GAUGE", ValueType: "BOOL"},
	"node_pool_status":                                                 {Kind: "GAUGE", ValueType: "BOOL"},
	"pod_ephemeral_storage_used_bytes":                                 {Kind: "GAUGE", ValueType: "INT64"},
	"pod_latencies_pod_first_ready":                                    {Kind: "GAUGE", ValueType: "DOUBLE"},
	"pod_network_policy_event_count":                                   {Kind: "DELTA", ValueType: "INT64"},
	"pod_network_received_bytes_count":                                 {Kind: "CUMULATIVE", ValueType: "INT64"},
	"pod_network_sent_bytes_count":                                     {Kind: "CUMULATIVE", ValueType: "INT64"},
	"pod_volume_utilization":                                           {Kind: "GAUGE", ValueType: "DOUBLE"},
}

var (
	promqlBlockRe  = regexp.MustCompile("(?s)```promql\\s*(.*?)\\s*```")
	metricRe       = regexp.MustCompile(`kubernetes_io:([a-zA-Z0-9_]+)`)
	rateCallRe     = regexp.MustCompile(`(?s)(?:rate|increase)\s*\(\s*[^)]*?kubernetes_io:([a-zA-Z0-9_]+)`)
	bucketMetricRe = regexp.MustCompile(`kubernetes_io:([a-zA-Z0-9_]+)_bucket`)
)

func TestSampleQueriesMetricSemantics(t *testing.T) {
	entries, err := os.ReadDir("schemas")
	if err != nil {
		t.Fatalf("read schemas directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join("schemas", entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read schema %s: %v", path, err)
		}

		blocks := promqlBlockRe.FindAllStringSubmatch(string(content), -1)
		for _, block := range blocks {
			query := block[1]
			validateQueryMetrics(t, path, query)
		}
	}
}

func validateQueryMetrics(t *testing.T, path, query string) {
	hasKubernetesMetric := false
	for _, match := range metricRe.FindAllStringSubmatch(query, -1) {
		metric := normalizeMetric(match[1])
		hasKubernetesMetric = true
		if _, ok := metricCatalog[metric]; !ok {
			t.Errorf("%s: unknown metric %q in sample query", path, metric)
		}
	}

	for _, match := range rateCallRe.FindAllStringSubmatch(query, -1) {
		metric := normalizeMetric(match[1])
		info, ok := metricCatalog[metric]
		if !ok {
			continue
		}
		if info.Kind != "CUMULATIVE" && info.Kind != "DELTA" {
			t.Errorf("%s: rate/increase used with %q (kind=%s)", path, metric, info.Kind)
		}
	}

	if strings.Contains(query, "histogram_quantile") && hasKubernetesMetric {
		buckets := bucketMetricRe.FindAllStringSubmatch(query, -1)
		if len(buckets) == 0 {
			t.Errorf("%s: histogram_quantile used without _bucket metric", path)
			return
		}
		for _, match := range buckets {
			metric := normalizeMetric(match[1])
			info, ok := metricCatalog[metric]
			if !ok {
				continue
			}
			if info.ValueType != "DISTRIBUTION" {
				t.Errorf("%s: histogram_quantile used with %q (type=%s)", path, metric, info.ValueType)
			}
		}
	}
}

func normalizeMetric(name string) string {
	if strings.HasSuffix(name, "_bucket") {
		return strings.TrimSuffix(name, "_bucket")
	}
	return name
}
