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

// Package controller2 implements client-go-based controllers for the insights
// aggregated API server. This package does not use controller-runtime.
package controller2

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/cel-go/cel"
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

	insightcel "github.com/datum-cloud/insights/internal/cel"
	"github.com/datum-cloud/insights/internal/metrics"
	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
	insightsclient "github.com/datum-cloud/insights/pkg/client/clientset"
)

const (
	insightPolicyFinalizer = "insights.miloapis.com/policy-finalizer"

	labelPolicyName = "insights.miloapis.com/policy-name"
	labelPolicyRule = "insights.miloapis.com/policy-rule"

	defaultPageSize = 500

	// Condition types for InsightPolicy
	conditionTypePolicyReady = "Ready"
	conditionTypePolicyValid = "Valid"
)

// compiledRule holds a compiled CEL condition and templates for a rule.
type compiledRule struct {
	name            string
	condition       cel.Program
	messageTemplate *insightcel.CompiledTemplate
	descriptionTmpl *insightcel.CompiledTemplate
	severity        insightsv1alpha1.InsightSeverity
	category        string
	ttlSeconds      int64
}

// PolicyController reconciles InsightPolicy resources using client-go informers.
type PolicyController struct {
	insightsClient insightsclient.Interface
	dynamicClient  dynamic.Interface
	recorder       record.EventRecorder

	policyInformer cache.SharedIndexInformer
	insightLister  cache.GenericLister

	queue workqueue.TypedRateLimitingInterface[string]
}

// NewPolicyController creates a new PolicyController.
func NewPolicyController(
	insightsClient insightsclient.Interface,
	dynamicClient dynamic.Interface,
	policyInformer cache.SharedIndexInformer,
	insightInformer cache.SharedIndexInformer,
	recorder record.EventRecorder,
) *PolicyController {
	c := &PolicyController{
		insightsClient: insightsClient,
		dynamicClient:  dynamicClient,
		recorder:       recorder,
		policyInformer: policyInformer,
		insightLister:  cache.NewGenericLister(insightInformer.GetIndexer(), insightsv1alpha1.Resource("insights")),
		queue: workqueue.NewTypedRateLimitingQueue(
			workqueue.DefaultTypedControllerRateLimiter[string](),
		),
	}

	_, _ = policyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
func (c *PolicyController) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	klog.Info("starting policy controller")
	defer klog.Info("stopping policy controller")

	if !cache.WaitForCacheSync(stopCh, c.policyInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
}

func (c *PolicyController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *PolicyController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.syncPolicy(context.Background(), key)
	if err == nil {
		c.queue.Forget(key)
		return true
	}

	runtime.HandleError(fmt.Errorf("syncing policy %q failed: %w", key, err))
	c.queue.AddRateLimited(key)
	return true
}

func (c *PolicyController) syncPolicy(ctx context.Context, key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	startTime := time.Now()

	policy, err := c.insightsClient.InsightsV1alpha1().InsightPolicies().Get(ctx, name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get InsightPolicy %q: %w", name, err)
	}

	// Handle deletion with finalizer
	if !policy.DeletionTimestamp.IsZero() {
		return c.handleDeletion(ctx, policy)
	}

	// Ensure finalizer
	if !containsFinalizer(policy.Finalizers, insightPolicyFinalizer) {
		return c.addFinalizer(ctx, policy)
	}

	// Check if suspended
	if policy.Spec.Suspended {
		return c.handleSuspended(ctx, policy)
	}

	// Compile rules
	compiledRules, err := compileRules(policy)
	if err != nil {
		klog.Errorf("failed to compile rules for policy %q: %v", policy.Name, err)
		metrics.InsightPolicyErrors.WithLabelValues(policy.Name, "compilation").Inc()
		return c.updatePolicyStatus(ctx, policy, func(status *insightsv1alpha1.InsightPolicyStatus) {
			status.Phase = insightsv1alpha1.InsightPolicyPhaseError
			setCondition(&status.Conditions, conditionTypePolicyValid, metav1.ConditionFalse,
				"CompilationError",
				fmt.Sprintf("Failed to compile policy rules: %v", err),
				policy.Generation)
		})
	}

	setConditionInPlace := func(status *insightsv1alpha1.InsightPolicyStatus) {
		setCondition(&status.Conditions, conditionTypePolicyValid, metav1.ConditionTrue,
			"Valid",
			fmt.Sprintf("All %d rules compiled successfully", len(compiledRules)),
			policy.Generation)
	}
	_ = setConditionInPlace

	// Compile match expression if present
	var matchProgram cel.Program
	if policy.Spec.Selector.MatchExpression != "" {
		matchProgram, err = insightcel.CompileCondition(policy.Spec.Selector.MatchExpression)
		if err != nil {
			klog.Errorf("failed to compile matchExpression for policy %q: %v", policy.Name, err)
			metrics.InsightPolicyErrors.WithLabelValues(policy.Name, "match_expression").Inc()
			if c.recorder != nil {
				c.recorder.Eventf(policy, "Warning", "EvaluationError",
					"Invalid matchExpression in selector: %v", err)
			}
			return c.updatePolicyStatus(ctx, policy, func(status *insightsv1alpha1.InsightPolicyStatus) {
				status.Phase = insightsv1alpha1.InsightPolicyPhaseError
				setCondition(&status.Conditions, conditionTypePolicyValid, metav1.ConditionFalse,
					"MatchExpressionError",
					fmt.Sprintf("Invalid matchExpression in selector: %v", err),
					policy.Generation)
			})
		}
	}

	// List matching resources
	resources, skippedCount, err := c.listMatchingResources(ctx, policy, matchProgram)
	if err != nil {
		klog.Errorf("failed to list matching resources for policy %q: %v", policy.Name, err)
		metrics.InsightPolicyErrors.WithLabelValues(policy.Name, "resource_listing").Inc()
		return err
	}

	// Evaluate rules and create/update insights
	ruleStatuses, totalInsightCount := c.evaluateRules(ctx, policy, resources, compiledRules)

	// Cleanup stale insights
	if err := c.cleanupStaleInsights(ctx, policy, resources, compiledRules); err != nil {
		klog.Errorf("failed to cleanup stale insights for policy %q: %v", policy.Name, err)
		// Continue, don't fail the reconciliation
	}

	// Record duration metric
	duration := time.Since(startTime).Seconds()
	metrics.InsightPolicyEvaluationDuration.Observe(duration)
	metrics.InsightPolicyResourcesMatched.WithLabelValues(policy.Name).Set(float64(len(resources)))

	// Determine whether the policy was previously active
	wasActive := policy.Status.Phase == insightsv1alpha1.InsightPolicyPhaseActive

	// Build status message
	readyMsg := fmt.Sprintf("Successfully evaluated %d resources and generated %d insights",
		len(resources), totalInsightCount)
	if skippedCount > 0 {
		readyMsg = fmt.Sprintf("Evaluated %d resources (%d skipped due to evaluation errors), generated %d insights",
			len(resources), skippedCount, totalInsightCount)
	}

	statusErr := c.updatePolicyStatus(ctx, policy, func(status *insightsv1alpha1.InsightPolicyStatus) {
		status.Phase = insightsv1alpha1.InsightPolicyPhaseActive
		status.MatchingResourceCount = int32(len(resources))
		status.SkippedResourceCount = skippedCount
		status.TotalInsightCount = totalInsightCount
		status.RuleStatuses = ruleStatuses
		now := metav1.Now()
		status.LastEvaluationTime = &now
		status.ObservedGeneration = policy.Generation
		setCondition(&status.Conditions, conditionTypePolicyReady, metav1.ConditionTrue,
			"Active", readyMsg, policy.Generation)
		setCondition(&status.Conditions, conditionTypePolicyValid, metav1.ConditionTrue,
			"Valid",
			fmt.Sprintf("All %d rules compiled successfully", len(compiledRules)),
			policy.Generation)
	})
	if statusErr != nil {
		return statusErr
	}

	// Emit events
	if c.recorder != nil {
		if !wasActive {
			c.recorder.Event(policy, "Normal", "PolicyActivated",
				"Policy has been activated and started evaluating resources")
		}
		c.recorder.Eventf(policy, "Normal", "InsightsGenerated",
			"Generated %d insights from %d resources", totalInsightCount, len(resources))
	}

	// Requeue after resync period
	resyncPeriod := time.Duration(policy.Spec.ResyncPeriodSeconds) * time.Second
	if resyncPeriod == 0 {
		resyncPeriod = 5 * time.Minute
	}
	c.queue.AddAfter(policy.Name, resyncPeriod)

	return nil
}

func (c *PolicyController) handleDeletion(ctx context.Context, policy *insightsv1alpha1.InsightPolicy) error {
	if !containsFinalizer(policy.Finalizers, insightPolicyFinalizer) {
		return nil
	}

	klog.Infof("cleaning up insights created by policy %q", policy.Name)
	if err := c.deleteAllPolicyInsights(ctx, policy); err != nil {
		return fmt.Errorf("failed to delete policy insights: %w", err)
	}

	return c.removeFinalizer(ctx, policy)
}

func (c *PolicyController) handleSuspended(ctx context.Context, policy *insightsv1alpha1.InsightPolicy) error {
	if policy.Status.Phase != insightsv1alpha1.InsightPolicyPhaseSuspended && c.recorder != nil {
		c.recorder.Event(policy, "Normal", "PolicySuspended", "Policy evaluation has been suspended")
	}
	return c.updatePolicyStatus(ctx, policy, func(status *insightsv1alpha1.InsightPolicyStatus) {
		status.Phase = insightsv1alpha1.InsightPolicyPhaseSuspended
		status.ObservedGeneration = policy.Generation
		setCondition(&status.Conditions, conditionTypePolicyReady, metav1.ConditionFalse,
			"Suspended",
			"Policy evaluation is suspended. Set spec.suspended to false to resume.",
			policy.Generation)
	})
}

func (c *PolicyController) addFinalizer(ctx context.Context, policy *insightsv1alpha1.InsightPolicy) error {
	patch := []byte(fmt.Sprintf(`{"metadata":{"finalizers":[%q]}}`, insightPolicyFinalizer))
	_, err := c.insightsClient.InsightsV1alpha1().InsightPolicies().Patch(
		ctx, policy.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *PolicyController) removeFinalizer(ctx context.Context, policy *insightsv1alpha1.InsightPolicy) error {
	// Build the new finalizers list without our finalizer
	newFinalizers := make([]string, 0, len(policy.Finalizers))
	for _, f := range policy.Finalizers {
		if f != insightPolicyFinalizer {
			newFinalizers = append(newFinalizers, f)
		}
	}

	finalizersJSON, err := json.Marshal(newFinalizers)
	if err != nil {
		return fmt.Errorf("failed to marshal finalizers: %w", err)
	}
	patch := []byte(fmt.Sprintf(`{"metadata":{"finalizers":%s}}`, finalizersJSON))
	_, err = c.insightsClient.InsightsV1alpha1().InsightPolicies().Patch(
		ctx, policy.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	return err
}

func (c *PolicyController) updatePolicyStatus(ctx context.Context, policy *insightsv1alpha1.InsightPolicy, mutate func(*insightsv1alpha1.InsightPolicyStatus)) error {
	// Work on a copy to avoid mutating the cached object
	updated := policy.DeepCopy()
	mutate(&updated.Status)

	_, err := c.insightsClient.InsightsV1alpha1().InsightPolicies().UpdateStatus(ctx, updated, metav1.UpdateOptions{})
	return err
}

func (c *PolicyController) listMatchingResources(ctx context.Context, policy *insightsv1alpha1.InsightPolicy, matchProgram cel.Program) ([]*unstructured.Unstructured, int32, error) {
	selector := policy.Spec.Selector

	gv, err := schema.ParseGroupVersion(selector.APIVersion)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid apiVersion %q: %w", selector.APIVersion, err)
	}

	// Determine resource name from kind (lowercase plural)
	resource := strings.ToLower(selector.Kind) + "s"
	gvr := schema.GroupVersionResource{
		Group:    gv.Group,
		Version:  gv.Version,
		Resource: resource,
	}

	var listOpts metav1.ListOptions
	listOpts.Limit = defaultPageSize

	if selector.LabelSelector != nil {
		labelSelector, err := metav1.LabelSelectorAsSelector(selector.LabelSelector)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid label selector: %w", err)
		}
		listOpts.LabelSelector = labelSelector.String()
	}

	namespaces := selector.Namespaces
	if len(namespaces) == 0 {
		namespaces = []string{""}
	}

	var allResources []*unstructured.Unstructured
	var skippedCount int32

	for _, ns := range namespaces {
		var resourceClient dynamic.ResourceInterface
		if ns == "" {
			resourceClient = c.dynamicClient.Resource(gvr)
		} else {
			resourceClient = c.dynamicClient.Resource(gvr).Namespace(ns)
		}

		continueToken := ""
		for {
			opts := listOpts
			opts.Continue = continueToken

			list, err := resourceClient.List(ctx, opts)
			if err != nil {
				nsDisplay := ns
				if nsDisplay == "" {
					nsDisplay = "all namespaces"
				}
				return nil, 0, fmt.Errorf("failed to list %s resources in %s: %w", selector.Kind, nsDisplay, err)
			}

			for i := range list.Items {
				item := &list.Items[i]

				if matchProgram != nil {
					matches, err := insightcel.EvaluateCondition(matchProgram, item)
					if err != nil {
						klog.V(1).Infof("skipped resource %s/%s due to match expression error: %v",
							item.GetNamespace(), item.GetName(), err)
						skippedCount++
						continue
					}
					if !matches {
						continue
					}
				}

				allResources = append(allResources, item)
			}

			continueToken = list.GetContinue()
			if continueToken == "" {
				break
			}
		}
	}

	return allResources, skippedCount, nil
}

func (c *PolicyController) evaluateRules(ctx context.Context, policy *insightsv1alpha1.InsightPolicy, resources []*unstructured.Unstructured, rules []compiledRule) ([]insightsv1alpha1.InsightRuleStatus, int32) {
	ruleStatuses := make([]insightsv1alpha1.InsightRuleStatus, len(rules))
	var totalInsightCount int32

	for i, rule := range rules {
		ruleStatuses[i] = insightsv1alpha1.InsightRuleStatus{
			Name:       rule.name,
			MatchCount: int32(len(resources)),
		}

		for _, obj := range resources {
			matches, err := insightcel.EvaluateCondition(rule.condition, obj)
			if err != nil {
				klog.Errorf("failed to evaluate condition for rule %q on resource %s/%s: %v",
					rule.name, obj.GetNamespace(), obj.GetName(), err)
				metrics.InsightPolicyErrors.WithLabelValues(policy.Name, "evaluation").Inc()
				ruleStatuses[i].Error = err.Error()
				continue
			}

			if !matches {
				continue
			}

			insight, err := generateInsight(policy, rule, obj)
			if err != nil {
				klog.Errorf("failed to generate insight for rule %q on resource %s/%s: %v",
					rule.name, obj.GetNamespace(), obj.GetName(), err)
				metrics.InsightPolicyErrors.WithLabelValues(policy.Name, "insight_generation").Inc()
				ruleStatuses[i].Error = err.Error()
				continue
			}

			if err := c.createOrUpdateInsight(ctx, insight); err != nil {
				klog.Errorf("failed to create/update insight %q: %v", insight.Name, err)
				metrics.InsightPolicyErrors.WithLabelValues(policy.Name, "insight_creation").Inc()
				continue
			}

			ruleStatuses[i].InsightCount++
			totalInsightCount++
		}
	}

	return ruleStatuses, totalInsightCount
}

func (c *PolicyController) createOrUpdateInsight(ctx context.Context, insight *insightsv1alpha1.Insight) error {
	existing, err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Get(
		ctx, insight.Name, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Create(
			ctx, insight, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	// Preserve status; only update spec and labels
	updated := existing.DeepCopy()
	updated.Spec = insight.Spec
	updated.Labels = insight.Labels

	_, err = c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Update(
		ctx, updated, metav1.UpdateOptions{})
	return err
}

func (c *PolicyController) cleanupStaleInsights(ctx context.Context, policy *insightsv1alpha1.InsightPolicy, resources []*unstructured.Unstructured, rules []compiledRule) error {
	// Build set of expected insight names
	expected := make(map[string]struct{})
	for _, rule := range rules {
		for _, obj := range resources {
			matches, err := insightcel.EvaluateCondition(rule.condition, obj)
			if err != nil || !matches {
				continue
			}
			name := generateInsightName(policy, rule.name, obj)
			expected[name] = struct{}{}
		}
	}

	// List insights by policy label across all namespaces
	list, err := c.insightsClient.InsightsV1alpha1().Insights("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelPolicyName, policy.Name),
	})
	if err != nil {
		return fmt.Errorf("failed to list insights for policy %q: %w", policy.Name, err)
	}

	for _, insight := range list.Items {
		if _, ok := expected[insight.Name]; !ok {
			err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Delete(
				ctx, insight.Name, metav1.DeleteOptions{})
			if err != nil && !errors.IsNotFound(err) {
				klog.Errorf("failed to delete stale insight %s/%s: %v",
					insight.Namespace, insight.Name, err)
			}
		}
	}

	return nil
}

func (c *PolicyController) deleteAllPolicyInsights(ctx context.Context, policy *insightsv1alpha1.InsightPolicy) error {
	list, err := c.insightsClient.InsightsV1alpha1().Insights("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", labelPolicyName, policy.Name),
	})
	if err != nil {
		return fmt.Errorf("failed to list insights for policy %q: %w", policy.Name, err)
	}

	for _, insight := range list.Items {
		err := c.insightsClient.InsightsV1alpha1().Insights(insight.Namespace).Delete(
			ctx, insight.Name, metav1.DeleteOptions{})
		if err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete insight %s/%s: %w", insight.Namespace, insight.Name, err)
		}
	}

	return nil
}

// compileRules compiles all CEL conditions and templates from the policy spec.
func compileRules(policy *insightsv1alpha1.InsightPolicy) ([]compiledRule, error) {
	rules := make([]compiledRule, 0, len(policy.Spec.Rules))

	for _, rule := range policy.Spec.Rules {
		condition, err := insightcel.CompileCondition(rule.Condition)
		if err != nil {
			return nil, fmt.Errorf("rule %q: invalid condition: %w", rule.Name, err)
		}

		msgTmpl, err := insightcel.CompileTemplate(rule.MessageTemplate)
		if err != nil {
			return nil, fmt.Errorf("rule %q: invalid messageTemplate: %w", rule.Name, err)
		}

		var descTmpl *insightcel.CompiledTemplate
		if rule.DescriptionTemplate != "" {
			descTmpl, err = insightcel.CompileTemplate(rule.DescriptionTemplate)
			if err != nil {
				return nil, fmt.Errorf("rule %q: invalid descriptionTemplate: %w", rule.Name, err)
			}
		}

		rules = append(rules, compiledRule{
			name:            rule.Name,
			condition:       condition,
			messageTemplate: msgTmpl,
			descriptionTmpl: descTmpl,
			severity:        rule.Severity,
			category:        rule.Category,
			ttlSeconds:      rule.TTLSeconds,
		})
	}

	return rules, nil
}

// generateInsight creates an Insight object from a matched resource.
func generateInsight(policy *insightsv1alpha1.InsightPolicy, rule compiledRule, obj *unstructured.Unstructured) (*insightsv1alpha1.Insight, error) {
	message, err := rule.messageTemplate.Evaluate(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate messageTemplate for %s/%s: %w",
			obj.GetKind(), obj.GetName(), err)
	}

	var description string
	if rule.descriptionTmpl != nil {
		description, err = rule.descriptionTmpl.Evaluate(obj)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate descriptionTemplate for %s/%s: %w",
				obj.GetKind(), obj.GetName(), err)
		}
	}

	insightName := generateInsightName(policy, rule.name, obj)

	insightNamespace := obj.GetNamespace()
	if insightNamespace == "" {
		insightNamespace = policy.Spec.InsightNamespace
	}

	return &insightsv1alpha1.Insight{
		ObjectMeta: metav1.ObjectMeta{
			Name:      insightName,
			Namespace: insightNamespace,
			Labels: map[string]string{
				labelPolicyName: policy.Name,
				labelPolicyRule: rule.name,
			},
		},
		Spec: insightsv1alpha1.InsightSpec{
			TargetRef: insightsv1alpha1.TargetReference{
				APIVersion: obj.GetAPIVersion(),
				Kind:       obj.GetKind(),
				Name:       obj.GetName(),
				Namespace:  obj.GetNamespace(),
			},
			Severity:    rule.severity,
			Message:     strings.TrimSpace(message),
			Description: strings.TrimSpace(description),
			Category:    rule.category,
			Source: insightsv1alpha1.InsightSource{
				Type: insightsv1alpha1.InsightSourceTypePolicy,
				PolicyRef: &insightsv1alpha1.PolicyReference{
					Name:     policy.Name,
					RuleName: rule.name,
				},
			},
			TTLSeconds: rule.ttlSeconds,
		},
	}, nil
}

// generateInsightName produces a deterministic name for an insight.
func generateInsightName(policy *insightsv1alpha1.InsightPolicy, ruleName string, obj *unstructured.Unstructured) string {
	prefix := policy.Spec.InsightNamePrefix
	if prefix == "" {
		prefix = policy.Name
	}

	hashInput := fmt.Sprintf("%s/%s/%s/%s/%s",
		policy.Name, ruleName,
		obj.GetNamespace(), obj.GetAPIVersion()+"/"+obj.GetKind(), obj.GetName())
	hash := sha256.Sum256([]byte(hashInput))
	shortHash := hex.EncodeToString(hash[:])[:8]

	name := fmt.Sprintf("%s-%s-%s", prefix, ruleName, shortHash)
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

// setCondition sets a condition on the conditions slice.
func setCondition(conditions *[]metav1.Condition, condType string, status metav1.ConditionStatus, reason, message string, generation int64) {
	now := metav1.Now()
	for i, c := range *conditions {
		if c.Type == condType {
			if c.Status != status || c.Reason != reason || c.Message != message {
				(*conditions)[i].Status = status
				(*conditions)[i].Reason = reason
				(*conditions)[i].Message = message
				(*conditions)[i].LastTransitionTime = now
				(*conditions)[i].ObservedGeneration = generation
			}
			return
		}
	}
	*conditions = append(*conditions, metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
		ObservedGeneration: generation,
	})
}

// containsFinalizer checks whether a finalizer is in the list.
func containsFinalizer(finalizers []string, finalizer string) bool {
	for _, f := range finalizers {
		if f == finalizer {
			return true
		}
	}
	return false
}
