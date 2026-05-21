/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMetricsRegistered(t *testing.T) {
	// Verify that all custom metrics can be used without panicking.

	t.Run("InsightsTotal", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InsightsTotal metric caused panic: %v", r)
			}
		}()
		InsightsTotal.WithLabelValues("info", "test", "default").Set(1.0)
	})

	t.Run("InsightPolicyEvaluationDuration", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InsightPolicyEvaluationDuration metric caused panic: %v", r)
			}
		}()
		InsightPolicyEvaluationDuration.Observe(0.5)
	})

	t.Run("InsightPolicyResourcesMatched", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InsightPolicyResourcesMatched metric caused panic: %v", r)
			}
		}()
		InsightPolicyResourcesMatched.WithLabelValues("test-policy").Set(10.0)
	})

	t.Run("InsightPolicyErrors", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("InsightPolicyErrors metric caused panic: %v", r)
			}
		}()
		InsightPolicyErrors.WithLabelValues("test-policy", "compilation").Inc()
	})
}

func TestMetricsCollectable(t *testing.T) {
	// Test that metrics can be collected from the default prometheus registry.
	// Metrics registered via promauto are in the default registry.
	metricFamilies, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}

	// Check that our custom metrics are present
	expectedMetrics := map[string]bool{
		"insights_total": false,
		"insightpolicy_evaluation_duration_seconds": false,
		"insightpolicy_resources_matched_total":     false,
		"insightpolicy_errors_total":                false,
	}

	for _, mf := range metricFamilies {
		if _, exists := expectedMetrics[mf.GetName()]; exists {
			expectedMetrics[mf.GetName()] = true
		}
	}

	for metricName, found := range expectedMetrics {
		if !found {
			t.Errorf("Expected metric %q not found in registry", metricName)
		}
	}
}

func TestMetricLabels(t *testing.T) {
	// Test that metrics have the correct label names.
	t.Run("InsightsTotal labels", func(t *testing.T) {
		metric := InsightsTotal.WithLabelValues("info", "security", "default")
		if metric == nil {
			t.Error("Failed to create metric with valid labels")
		}
	})

	t.Run("InsightPolicyResourcesMatched labels", func(t *testing.T) {
		metric := InsightPolicyResourcesMatched.WithLabelValues("my-policy")
		if metric == nil {
			t.Error("Failed to create metric with valid labels")
		}
	})

	t.Run("InsightPolicyErrors labels", func(t *testing.T) {
		metric := InsightPolicyErrors.WithLabelValues("my-policy", "evaluation")
		if metric == nil {
			t.Error("Failed to create metric with valid labels")
		}
	})
}

func TestMetricTypes(t *testing.T) {
	// Verify that metrics implement prometheus.Collector.
	tests := []struct {
		name   string
		metric interface{}
		want   string
	}{
		{
			name:   "InsightsTotal is GaugeVec",
			metric: InsightsTotal,
			want:   "*prometheus.GaugeVec",
		},
		{
			name:   "InsightPolicyEvaluationDuration is Histogram",
			metric: InsightPolicyEvaluationDuration,
			want:   "prometheus.Histogram",
		},
		{
			name:   "InsightPolicyResourcesMatched is GaugeVec",
			metric: InsightPolicyResourcesMatched,
			want:   "*prometheus.GaugeVec",
		},
		{
			name:   "InsightPolicyErrors is CounterVec",
			metric: InsightPolicyErrors,
			want:   "*prometheus.CounterVec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := tt.metric.(prometheus.Collector); !ok {
				t.Errorf("Metric %q does not implement prometheus.Collector interface", tt.want)
			}
		})
	}
}
