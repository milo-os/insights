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

package controller2

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/datum-cloud/insights/internal/metrics"
	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
	insightsclient "github.com/datum-cloud/insights/pkg/client/clientset"
)

const (
	insightFinalizer = "insights.miloapis.com/finalizer"

	conditionTypeReady        = "Ready"
	conditionTypeTargetExists = "TargetExists"

	defaultRequeueInterval = 5 * time.Minute
)

// InsightController reconciles Insight resources using client-go informers.
type InsightController struct {
	insightsClient insightsclient.Interface
	dynamicClient  dynamic.Interface
	recorder       record.EventRecorder

	insightInformer cache.SharedIndexInformer

	queue workqueue.TypedRateLimitingInterface[string]
}

// NewInsightController creates a new InsightController.
func NewInsightController(
	insightsClient insightsclient.Interface,
	dynamicClient dynamic.Interface,
	insightInformer cache.SharedIndexInformer,
	recorder record.EventRecorder,
) *InsightController {
	c := &InsightController{
		insightsClient:  insightsClient,
		dynamicClient:   dynamicClient,
		recorder:        recorder,
		insightInformer: insightInformer,
		queue: workqueue.NewTypedRateLimitingQueue(
			workqueue.DefaultTypedControllerRateLimiter[string](),
		),
	}

	_, _ = insightInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err == nil {
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
	})

	return c
}

// Run starts the controller's worker goroutines.
func (c *InsightController) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("starting insight controller")
	defer klog.Info("stopping insight controller")

	if !cache.WaitForCacheSync(stopCh, c.insightInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *InsightController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *InsightController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	requeueAfter, err := c.syncInsight(context.Background(), key)
	if err == nil {
		c.queue.Forget(key)
		if requeueAfter > 0 {
			c.queue.AddAfter(key, requeueAfter)
		}
		return true
	}

	runtime.HandleError(fmt.Errorf("syncing insight %q failed: %w", key, err))
	c.queue.AddRateLimited(key)
	return true
}

func (c *InsightController) syncInsight(ctx context.Context, key string) (time.Duration, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return 0, err
	}

	insight, err := c.insightsClient.InsightsV1alpha1().Insights(namespace).Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get Insight %s/%s: %w", namespace, name, err)
	}

	// Handle deletion
	if !insight.DeletionTimestamp.IsZero() {
		return 0, c.handleInsightDeletion(ctx, insight)
	}

	// Ensure finalizer
	if !containsFinalizer(insight.Finalizers, insightFinalizer) {
		return 0, c.addInsightFinalizer(ctx, insight)
	}

	// Work on a copy
	updated := insight.DeepCopy()

	// Initialize state if empty
	if updated.Status.State == "" {
		updated.Status.State = insightsv1alpha1.InsightStateActive
		if c.recorder != nil {
			c.recorder.Event(insight, "Normal", "InsightCreated",
				"Insight has been created and is now active")
		}
	}

	// Handle snooze expiration: if snoozed and snooze.until is in the past, revert to Active
	if updated.Status.State == insightsv1alpha1.InsightStateSnoozed &&
		updated.Status.Snooze != nil &&
		time.Now().After(updated.Status.Snooze.Until.Time) {
		updated.Status.State = insightsv1alpha1.InsightStateActive
		updated.Status.Snooze = nil
		if c.recorder != nil {
			c.recorder.Event(insight, "Normal", "SnoozeExpired",
				"Snooze period has expired; insight is now active")
		}
	}

	// Set expiration time if TTL configured and not yet set
	if updated.Spec.TTLSeconds > 0 && updated.Status.ExpiresAt == nil {
		expiresAt := metav1.NewTime(insight.CreationTimestamp.Add(
			time.Duration(updated.Spec.TTLSeconds) * time.Second))
		updated.Status.ExpiresAt = &expiresAt
	}

	// Check TTL expiration
	if updated.Status.ExpiresAt != nil && time.Now().After(updated.Status.ExpiresAt.Time) {
		return 0, c.handleInsightExpiration(ctx, insight)
	}

	// Check resolved deleteAt
	if updated.Status.DeleteAt != nil && time.Now().After(updated.Status.DeleteAt.Time) {
		return 0, c.handleInsightExpiration(ctx, insight)
	}

	// Check target existence
	targetExists, err := c.checkTargetExists(ctx, insight)
	if err != nil {
		klog.Errorf("failed to check target existence for insight %s/%s: %v",
			namespace, name, err)
		setCondition(&updated.Status.Conditions, conditionTypeTargetExists,
			metav1.ConditionUnknown, "CheckFailed",
			fmt.Sprintf("Unable to verify target existence: %v", err),
			insight.Generation)
	} else {
		updated.Status.TargetExists = &targetExists
		if targetExists {
			setCondition(&updated.Status.Conditions, conditionTypeTargetExists,
				metav1.ConditionTrue, "TargetFound",
				fmt.Sprintf("Target resource %s/%s exists", insight.Spec.TargetRef.Kind, insight.Spec.TargetRef.Name),
				insight.Generation)
		} else {
			setCondition(&updated.Status.Conditions, conditionTypeTargetExists,
				metav1.ConditionFalse, "TargetNotFound",
				fmt.Sprintf("Target resource %s/%s was not found", insight.Spec.TargetRef.Kind, insight.Spec.TargetRef.Name),
				insight.Generation)
			if c.recorder != nil {
				c.recorder.Eventf(insight, "Warning", "TargetNotFound",
					"Target resource %s/%s not found", insight.Spec.TargetRef.Kind, insight.Spec.TargetRef.Name)
			}
		}
	}

	// Set Ready condition based on state
	if updated.Status.State == insightsv1alpha1.InsightStateActive {
		setCondition(&updated.Status.Conditions, conditionTypeReady,
			metav1.ConditionTrue, "Active",
			"Insight is active and monitoring the target resource",
			insight.Generation)
	}

	updated.Status.ObservedGeneration = insight.Generation

	// Update status
	_, statusErr := c.insightsClient.InsightsV1alpha1().Insights(namespace).UpdateStatus(
		ctx, updated, metav1.UpdateOptions{})
	if statusErr != nil {
		return 0, fmt.Errorf("failed to update Insight status for %s/%s: %w", namespace, name, statusErr)
	}

	// Update metrics
	c.updateInsightMetrics(ctx)

	return c.calculateRequeueAfter(updated), nil
}

func (c *InsightController) handleInsightDeletion(ctx context.Context, insight *insightsv1alpha1.Insight) error {
	if !containsFinalizer(insight.Finalizers, insightFinalizer) {
		return nil
	}

	klog.Infof("removing finalizer from insight %s/%s", insight.Namespace, insight.Name)

	// Build new finalizers list
	newFinalizers := make([]string, 0, len(insight.Finalizers))
	for _, f := range insight.Finalizers {
		if f != insightFinalizer {
			newFinalizers = append(newFinalizers, f)
		}
	}

	updated := insight.DeepCopy()
	updated.Finalizers = newFinalizers
	_, err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Update(
		ctx, updated, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove finalizer from insight %s/%s: %w",
			insight.Namespace, insight.Name, err)
	}

	c.updateInsightMetrics(ctx)
	return nil
}

func (c *InsightController) handleInsightExpiration(ctx context.Context, insight *insightsv1alpha1.Insight) error {
	klog.Infof("deleting expired insight %s/%s", insight.Namespace, insight.Name)
	if c.recorder != nil {
		c.recorder.Eventf(insight, "Normal", "InsightExpired",
			"Insight expired after TTL of %d seconds", insight.Spec.TTLSeconds)
	}
	err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Delete(
		ctx, insight.Name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete expired insight %s/%s: %w",
			insight.Namespace, insight.Name, err)
	}
	return nil
}

func (c *InsightController) addInsightFinalizer(ctx context.Context, insight *insightsv1alpha1.Insight) error {
	patch := []byte(fmt.Sprintf(`{"metadata":{"finalizers":[%q]}}`, insightFinalizer))
	_, err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Patch(
		ctx, insight.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *InsightController) checkTargetExists(ctx context.Context, insight *insightsv1alpha1.Insight) (bool, error) {
	targetRef := insight.Spec.TargetRef

	gv, err := schema.ParseGroupVersion(targetRef.APIVersion)
	if err != nil {
		return false, fmt.Errorf("invalid apiVersion %q: %w", targetRef.APIVersion, err)
	}

	// Determine resource name from kind (lowercase plural)
	resource := fmt.Sprintf("%ss", targetRef.Kind)
	gvr := schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: resource,
	}

	var resourceClient dynamic.ResourceInterface
	if targetRef.Namespace != "" {
		resourceClient = c.dynamicClient.Resource(gvr).Namespace(targetRef.Namespace)
	} else {
		resourceClient = c.dynamicClient.Resource(gvr)
	}

	target := &unstructured.Unstructured{}
	target.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   gv.Group,
		Version: gv.Version,
		Kind:    targetRef.Kind,
	})

	_, err = resourceClient.Get(ctx, targetRef.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (c *InsightController) calculateRequeueAfter(insight *insightsv1alpha1.Insight) time.Duration {
	var earliest time.Time

	if insight.Status.ExpiresAt != nil {
		until := time.Until(insight.Status.ExpiresAt.Time)
		if until > 0 {
			if earliest.IsZero() || insight.Status.ExpiresAt.Time.Before(earliest) {
				earliest = insight.Status.ExpiresAt.Time
			}
		}
	}

	if insight.Status.DeleteAt != nil {
		until := time.Until(insight.Status.DeleteAt.Time)
		if until > 0 {
			if earliest.IsZero() || insight.Status.DeleteAt.Time.Before(earliest) {
				earliest = insight.Status.DeleteAt.Time
			}
		}
	}

	if insight.Status.Snooze != nil {
		until := time.Until(insight.Status.Snooze.Until.Time)
		if until > 0 {
			if earliest.IsZero() || insight.Status.Snooze.Until.Time.Before(earliest) {
				earliest = insight.Status.Snooze.Until.Time
			}
		}
	}

	if !earliest.IsZero() {
		d := time.Until(earliest)
		if d > 0 {
			return d
		}
		return time.Second
	}

	return defaultRequeueInterval
}

func (c *InsightController) updateInsightMetrics(ctx context.Context) {
	list, err := c.insightsClient.InsightsV1alpha1().Insights("").List(ctx, metav1.ListOptions{})
	if err != nil {
		klog.Errorf("failed to list insights for metrics: %v", err)
		return
	}

	countMap := make(map[string]map[string]map[string]int)
	for _, insight := range list.Items {
		if insight.Status.State == insightsv1alpha1.InsightStateResolved {
			continue
		}
		severity := string(insight.Spec.Severity)
		category := insight.Spec.Category
		namespace := insight.Namespace

		if countMap[severity] == nil {
			countMap[severity] = make(map[string]map[string]int)
		}
		if countMap[severity][category] == nil {
			countMap[severity][category] = make(map[string]int)
		}
		countMap[severity][category][namespace]++
	}

	metrics.InsightsTotal.Reset()
	for severity, categoryMap := range countMap {
		for category, namespaceMap := range categoryMap {
			for namespace, count := range namespaceMap {
				metrics.InsightsTotal.WithLabelValues(severity, category, namespace).Set(float64(count))
			}
		}
	}
}
