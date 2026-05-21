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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// InsightsTotal tracks the total number of active insights by severity, category, and namespace
	InsightsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "insights_total",
			Help: "Total number of active Insights, labeled by severity, category, and namespace",
		},
		[]string{"severity", "category", "namespace"},
	)

	// InsightPolicyEvaluationDuration tracks the time taken to evaluate a policy
	InsightPolicyEvaluationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "insightpolicy_evaluation_duration_seconds",
			Help:    "Time taken to evaluate an InsightPolicy in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// InsightPolicyResourcesMatched tracks the number of resources matched per policy
	InsightPolicyResourcesMatched = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "insightpolicy_resources_matched_total",
			Help: "Total number of resources matched by an InsightPolicy",
		},
		[]string{"policy"},
	)

	// InsightPolicyErrors tracks evaluation errors by policy and error type
	InsightPolicyErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "insightpolicy_errors_total",
			Help: "Total number of errors during InsightPolicy evaluation, labeled by policy and error type",
		},
		[]string{"policy", "error_type"},
	)
)
