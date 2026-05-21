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

// Package main is the entry point for the insights controller binary.
// This controller uses client-go directly, not controller-runtime.
package main

import (
	"context"
	"flag"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	controller2 "github.com/datum-cloud/insights/internal/controller2"
	_ "github.com/datum-cloud/insights/internal/metrics"
	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
	insightsclient "github.com/datum-cloud/insights/pkg/client/clientset"
)

func main() {
	var (
		kubeconfig     string
		apiServerURL   string
		metricsAddr    string
		policyWorkers  int
		insightWorkers int
	)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig file. If not set, in-cluster config is used.")
	flag.StringVar(&apiServerURL, "insights-apiserver-url", "",
		"URL of the insights aggregated API server. If empty, uses in-cluster config.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.IntVar(&policyWorkers, "policy-workers", 2, "Number of workers for the policy controller.")
	flag.IntVar(&insightWorkers, "insight-workers", 2, "Number of workers for the insight controller.")

	klog.InitFlags(nil)
	flag.Parse()

	// Build rest config for the cluster we're running against (for dynamic client)
	clusterConfig, err := buildRestConfig(kubeconfig)
	if err != nil {
		klog.Exitf("failed to build cluster rest config: %v", err)
	}

	// Build rest config for the insights API server
	insightsConfig := clusterConfig
	if apiServerURL != "" {
		insightsConfig = &rest.Config{
			Host:        apiServerURL,
			BearerToken: clusterConfig.BearerToken,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true, // Skip TLS verification for in-cluster aggregated API server
			},
		}
	}

	// Create insights typed client
	insightsClient, err := insightsclient.NewForConfig(insightsConfig)
	if err != nil {
		klog.Exitf("failed to create insights client: %v", err)
	}

	// Create dynamic client for querying arbitrary cluster resources
	dynamicClient, err := dynamic.NewForConfig(clusterConfig)
	if err != nil {
		klog.Exitf("failed to create dynamic client: %v", err)
	}

	// Create event recorder
	eventRecorder := buildEventRecorder()

	// Create stop channel
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	stopCh := ctx.Done()

	// Create informers for our resources
	insightInformer := buildInsightInformer(insightsClient)
	policyInformer := buildPolicyInformer(insightsClient)
	muteRuleInformer := buildMuteRuleInformer(insightsClient)

	// Create controllers
	policyCtrl := controller2.NewPolicyController(
		insightsClient,
		dynamicClient,
		policyInformer,
		insightInformer,
		eventRecorder,
	)

	insightCtrl := controller2.NewInsightController(
		insightsClient,
		dynamicClient,
		insightInformer,
		eventRecorder,
	)

	muteRuleCtrl := controller2.NewMuteRuleController(
		insightsClient,
		muteRuleInformer,
		eventRecorder,
	)

	// Start informers
	go insightInformer.Run(stopCh)
	go policyInformer.Run(stopCh)
	go muteRuleInformer.Run(stopCh)

	// Start metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})
		mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})
		srv := &http.Server{
			Addr:    metricsAddr,
			Handler: mux,
		}
		klog.Infof("starting metrics server at %s", metricsAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("metrics server error: %v", err)
		}
	}()

	// Start controllers
	go policyCtrl.Run(policyWorkers, stopCh)
	go insightCtrl.Run(insightWorkers, stopCh)
	go muteRuleCtrl.Run(1, stopCh)

	klog.Info("insights controller started")
	<-stopCh
	klog.Info("insights controller shutting down")
}

func buildRestConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	cfg, err := rest.InClusterConfig()
	if err == rest.ErrNotInCluster {
		// Fallback to default kubeconfig for local development
		return clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	}
	return cfg, err
}

func buildEventRecorder() record.EventRecorder {
	// Use a simple event recorder that logs events via structured logging.
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	return eventBroadcaster.NewRecorder(
		runtime.NewScheme(),
		corev1.EventSource{Component: "insights-controller"},
	)
}

// buildInsightInformer creates a SharedIndexInformer for Insight resources.
func buildInsightInformer(insightsClient insightsclient.Interface) cache.SharedIndexInformer {
	resyncPeriod := 10 * time.Minute
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return insightsClient.InsightsV1alpha1().Insights("").List(context.Background(), opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				opts.Watch = true
				opts.FieldSelector = fields.Everything().String()
				return insightsClient.InsightsV1alpha1().Insights("").Watch(context.Background(), opts)
			},
		},
		&insightsv1alpha1.Insight{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
}

// buildPolicyInformer creates a SharedIndexInformer for InsightPolicy resources.
func buildPolicyInformer(insightsClient insightsclient.Interface) cache.SharedIndexInformer {
	resyncPeriod := 10 * time.Minute
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return insightsClient.InsightsV1alpha1().InsightPolicies().List(context.Background(), opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				opts.Watch = true
				return insightsClient.InsightsV1alpha1().InsightPolicies().Watch(context.Background(), opts)
			},
		},
		&insightsv1alpha1.InsightPolicy{},
		resyncPeriod,
		cache.Indexers{},
	)
}

// buildMuteRuleInformer creates a SharedIndexInformer for InsightMuteRule resources.
func buildMuteRuleInformer(insightsClient insightsclient.Interface) cache.SharedIndexInformer {
	resyncPeriod := 10 * time.Minute
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return insightsClient.InsightsV1alpha1().InsightMuteRules("").List(context.Background(), opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				opts.Watch = true
				return insightsClient.InsightsV1alpha1().InsightMuteRules("").Watch(context.Background(), opts)
			},
		},
		&insightsv1alpha1.InsightMuteRule{},
		resyncPeriod,
		cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
	)
}
