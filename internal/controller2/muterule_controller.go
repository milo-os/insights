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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
	insightsclient "github.com/datum-cloud/insights/pkg/client/clientset"
)

const (
	muteRuleRequeueInterval = 5 * time.Minute
)

// MuteRuleController reconciles InsightMuteRule resources.
// It lists insights matching mute criteria and sets status.muted on them.
type MuteRuleController struct {
	insightsClient insightsclient.Interface
	recorder       record.EventRecorder

	muteRuleInformer cache.SharedIndexInformer

	queue workqueue.TypedRateLimitingInterface[string]
}

// NewMuteRuleController creates a new MuteRuleController.
func NewMuteRuleController(
	insightsClient insightsclient.Interface,
	muteRuleInformer cache.SharedIndexInformer,
	recorder record.EventRecorder,
) *MuteRuleController {
	c := &MuteRuleController{
		insightsClient:   insightsClient,
		recorder:         recorder,
		muteRuleInformer: muteRuleInformer,
		queue: workqueue.NewTypedRateLimitingQueue(
			workqueue.DefaultTypedControllerRateLimiter[string](),
		),
	}

	_, _ = muteRuleInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
func (c *MuteRuleController) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("starting muterule controller")
	defer klog.Info("stopping muterule controller")

	if !cache.WaitForCacheSync(stopCh, c.muteRuleInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *MuteRuleController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *MuteRuleController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	requeueAfter, err := c.syncMuteRule(context.Background(), key)
	if err == nil {
		c.queue.Forget(key)
		if requeueAfter > 0 {
			c.queue.AddAfter(key, requeueAfter)
		}
		return true
	}

	runtime.HandleError(fmt.Errorf("syncing mute rule %q failed: %w", key, err))
	c.queue.AddRateLimited(key)
	return true
}

func (c *MuteRuleController) syncMuteRule(ctx context.Context, key string) (time.Duration, error) {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return 0, err
	}

	muteRule, err := c.insightsClient.InsightsV1alpha1().InsightMuteRules(namespace).Get(
		ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get InsightMuteRule %s/%s: %w", namespace, name, err)
	}

	updated := muteRule.DeepCopy()
	updated.Status.ObservedGeneration = muteRule.Generation

	// Check expiration
	if muteRule.Spec.ExpiresAt != nil && time.Now().After(muteRule.Spec.ExpiresAt.Time) {
		updated.Status.Expired = true
		_, err := c.insightsClient.InsightsV1alpha1().InsightMuteRules(namespace).UpdateStatus(
			ctx, updated, metav1.UpdateOptions{})
		return 0, err
	}

	// List insights in the namespace matching criteria
	matchingInsights, err := c.findMatchingInsights(ctx, muteRule)
	if err != nil {
		return 0, fmt.Errorf("failed to find matching insights for mute rule %s/%s: %w",
			namespace, name, err)
	}

	// Update each matching insight
	mutedCount := int32(0)
	for _, insight := range matchingInsights {
		if insight.Status.Muted && insight.Status.MutedBy != nil &&
			insight.Status.MutedBy.Name == muteRule.Name &&
			insight.Status.MutedBy.Namespace == muteRule.Namespace {
			mutedCount++
			continue
		}

		toUpdate := insight.DeepCopy()
		toUpdate.Status.Muted = true
		toUpdate.Status.MutedBy = &insightsv1alpha1.MuteRuleReference{
			Name:      muteRule.Name,
			Namespace: muteRule.Namespace,
		}

		_, err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).UpdateStatus(
			ctx, toUpdate, metav1.UpdateOptions{})
		if err != nil {
			klog.Errorf("failed to update muted status for insight %s/%s: %v",
				insight.Namespace, insight.Name, err)
			continue
		}
		mutedCount++
	}

	// Update status
	updated.Status.MutedInsightCount = mutedCount
	_, statusErr := c.insightsClient.InsightsV1alpha1().InsightMuteRules(namespace).UpdateStatus(
		ctx, updated, metav1.UpdateOptions{})
	if statusErr != nil {
		return 0, fmt.Errorf("failed to update InsightMuteRule status %s/%s: %w",
			namespace, name, statusErr)
	}

	// Requeue at expiry or default interval
	if muteRule.Spec.ExpiresAt != nil {
		until := time.Until(muteRule.Spec.ExpiresAt.Time)
		if until > 0 {
			return until, nil
		}
	}

	return muteRuleRequeueInterval, nil
}

// findMatchingInsights lists all insights in the mute rule's namespace that match
// the mute rule's criteria.
func (c *MuteRuleController) findMatchingInsights(ctx context.Context, muteRule *insightsv1alpha1.InsightMuteRule) ([]*insightsv1alpha1.Insight, error) {
	match := muteRule.Spec.Match

	// Build list options; start with label selector if specified
	listOpts := metav1.ListOptions{}
	if match.LabelSelector != nil {
		sel, err := metav1.LabelSelectorAsSelector(match.LabelSelector)
		if err != nil {
			return nil, fmt.Errorf("invalid labelSelector: %w", err)
		}
		listOpts.LabelSelector = sel.String()
	}

	list, err := c.insightsClient.InsightsV1alpha1().Insights(muteRule.Namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	matched := make([]*insightsv1alpha1.Insight, 0, len(list.Items))
	for i := range list.Items {
		insight := &list.Items[i]

		if !insightMatchesMuteRule(insight, muteRule) {
			continue
		}

		matched = append(matched, insight)
	}

	return matched, nil
}

// insightMatchesMuteRule returns true if the insight matches all specified mute criteria.
func insightMatchesMuteRule(insight *insightsv1alpha1.Insight, muteRule *insightsv1alpha1.InsightMuteRule) bool {
	match := muteRule.Spec.Match

	// Already resolved insights should not be muted
	if insight.Status.State == insightsv1alpha1.InsightStateResolved {
		return false
	}

	// Filter by policyRef
	if match.PolicyRef != nil {
		if insight.Spec.Source.PolicyRef == nil {
			return false
		}
		if insight.Spec.Source.PolicyRef.Name != match.PolicyRef.Name {
			return false
		}
		if match.PolicyRef.RuleName != "" && insight.Spec.Source.PolicyRef.RuleName != match.PolicyRef.RuleName {
			return false
		}
	}

	// Filter by category
	if match.Category != "" && insight.Spec.Category != match.Category {
		return false
	}

	// Filter by targetRef
	if match.TargetRef != nil {
		t := insight.Spec.TargetRef
		mt := match.TargetRef
		if mt.Name != "" && t.Name != mt.Name {
			return false
		}
		if mt.Kind != "" && t.Kind != mt.Kind {
			return false
		}
		if mt.APIVersion != "" && t.APIVersion != mt.APIVersion {
			return false
		}
		if mt.Namespace != "" && t.Namespace != mt.Namespace {
			return false
		}
	}

	// Filter by severity (mute at-or-below specified severity)
	if match.Severity != nil {
		insightSev := severityOrdinal(insight.Spec.Severity)
		muteSev := severityOrdinal(*match.Severity)
		if insightSev > muteSev {
			return false
		}
	}

	// Filter by labelSelector
	if match.LabelSelector != nil {
		sel, err := metav1.LabelSelectorAsSelector(match.LabelSelector)
		if err != nil {
			return false
		}
		if !sel.Matches(labels.Set(insight.Labels)) {
			return false
		}
	}

	return true
}

// severityOrdinal returns a numeric value for severity comparison.
// Lower ordinal = less severe.
func severityOrdinal(s insightsv1alpha1.InsightSeverity) int {
	switch s {
	case insightsv1alpha1.InsightSeverityInfo:
		return 1
	case insightsv1alpha1.InsightSeverityWarning:
		return 2
	case insightsv1alpha1.InsightSeverityCritical:
		return 3
	default:
		return 0
	}
}
