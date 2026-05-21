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

package v1alpha1

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// InsightMuteRulesGetter has a method to return an InsightMuteRuleInterface.
type InsightMuteRulesGetter interface {
	InsightMuteRules(namespace string) InsightMuteRuleInterface
}

// InsightMuteRuleInterface has methods to work with InsightMuteRule resources.
type InsightMuteRuleInterface interface {
	Create(ctx context.Context, insightMuteRule *insightsv1alpha1.InsightMuteRule, opts metav1.CreateOptions) (*insightsv1alpha1.InsightMuteRule, error)
	Update(ctx context.Context, insightMuteRule *insightsv1alpha1.InsightMuteRule, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightMuteRule, error)
	UpdateStatus(ctx context.Context, insightMuteRule *insightsv1alpha1.InsightMuteRule, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightMuteRule, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*insightsv1alpha1.InsightMuteRule, error)
	List(ctx context.Context, opts metav1.ListOptions) (*insightsv1alpha1.InsightMuteRuleList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*insightsv1alpha1.InsightMuteRule, error)
}

// insightMuteRules implements InsightMuteRuleInterface.
type insightMuteRules struct {
	client rest.Interface
	ns     string
}

// newInsightMuteRules returns an insightMuteRules client.
func newInsightMuteRules(c *InsightsV1alpha1Client, namespace string) *insightMuteRules {
	return &insightMuteRules{client: c.RESTClient(), ns: namespace}
}

func (c *insightMuteRules) Get(ctx context.Context, name string, opts metav1.GetOptions) (*insightsv1alpha1.InsightMuteRule, error) {
	result := &insightsv1alpha1.InsightMuteRule{}
	err := c.client.Get().
		Namespace(c.ns).
		Resource("insightmuterules").
		Name(name).
		VersionedParams(&opts, parameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightMuteRules) List(ctx context.Context, opts metav1.ListOptions) (*insightsv1alpha1.InsightMuteRuleList, error) {
	result := &insightsv1alpha1.InsightMuteRuleList{}
	err := c.client.Get().
		Namespace(c.ns).
		Resource("insightmuterules").
		VersionedParams(&opts, parameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightMuteRules) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("insightmuterules").
		VersionedParams(&opts, parameterCodec).
		Watch(ctx)
}

func (c *insightMuteRules) Create(ctx context.Context, insightMuteRule *insightsv1alpha1.InsightMuteRule, opts metav1.CreateOptions) (*insightsv1alpha1.InsightMuteRule, error) {
	result := &insightsv1alpha1.InsightMuteRule{}
	err := c.client.Post().
		Namespace(c.ns).
		Resource("insightmuterules").
		VersionedParams(&opts, parameterCodec).
		Body(insightMuteRule).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightMuteRules) Update(ctx context.Context, insightMuteRule *insightsv1alpha1.InsightMuteRule, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightMuteRule, error) {
	result := &insightsv1alpha1.InsightMuteRule{}
	err := c.client.Put().
		Namespace(c.ns).
		Resource("insightmuterules").
		Name(insightMuteRule.Name).
		VersionedParams(&opts, parameterCodec).
		Body(insightMuteRule).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightMuteRules) UpdateStatus(ctx context.Context, insightMuteRule *insightsv1alpha1.InsightMuteRule, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightMuteRule, error) {
	result := &insightsv1alpha1.InsightMuteRule{}
	err := c.client.Put().
		Namespace(c.ns).
		Resource("insightmuterules").
		Name(insightMuteRule.Name).
		SubResource("status").
		VersionedParams(&opts, parameterCodec).
		Body(insightMuteRule).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightMuteRules) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("insightmuterules").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *insightMuteRules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*insightsv1alpha1.InsightMuteRule, error) {
	result := &insightsv1alpha1.InsightMuteRule{}
	err := c.client.Patch(pt).
		Namespace(c.ns).
		Resource("insightmuterules").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, parameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}
