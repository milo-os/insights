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

// InsightPoliciesGetter has a method to return an InsightPolicyInterface.
type InsightPoliciesGetter interface {
	InsightPolicies() InsightPolicyInterface
}

// InsightPolicyInterface has methods to work with InsightPolicy resources.
type InsightPolicyInterface interface {
	Create(ctx context.Context, insightPolicy *insightsv1alpha1.InsightPolicy, opts metav1.CreateOptions) (*insightsv1alpha1.InsightPolicy, error)
	Update(ctx context.Context, insightPolicy *insightsv1alpha1.InsightPolicy, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightPolicy, error)
	UpdateStatus(ctx context.Context, insightPolicy *insightsv1alpha1.InsightPolicy, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightPolicy, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*insightsv1alpha1.InsightPolicy, error)
	List(ctx context.Context, opts metav1.ListOptions) (*insightsv1alpha1.InsightPolicyList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*insightsv1alpha1.InsightPolicy, error)
}

// insightPolicies implements InsightPolicyInterface.
type insightPolicies struct {
	client rest.Interface
}

// newInsightPolicies returns an insightPolicies client.
func newInsightPolicies(c *InsightsV1alpha1Client) *insightPolicies {
	return &insightPolicies{client: c.RESTClient()}
}

func (c *insightPolicies) Get(ctx context.Context, name string, opts metav1.GetOptions) (*insightsv1alpha1.InsightPolicy, error) {
	result := &insightsv1alpha1.InsightPolicy{}
	err := c.client.Get().
		Resource("insightpolicies").
		Name(name).
		VersionedParams(&opts, parameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightPolicies) List(ctx context.Context, opts metav1.ListOptions) (*insightsv1alpha1.InsightPolicyList, error) {
	result := &insightsv1alpha1.InsightPolicyList{}
	err := c.client.Get().
		Resource("insightpolicies").
		VersionedParams(&opts, parameterCodec).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightPolicies) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("insightpolicies").
		VersionedParams(&opts, parameterCodec).
		Watch(ctx)
}

func (c *insightPolicies) Create(ctx context.Context, insightPolicy *insightsv1alpha1.InsightPolicy, opts metav1.CreateOptions) (*insightsv1alpha1.InsightPolicy, error) {
	result := &insightsv1alpha1.InsightPolicy{}
	err := c.client.Post().
		Resource("insightpolicies").
		VersionedParams(&opts, parameterCodec).
		Body(insightPolicy).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightPolicies) Update(ctx context.Context, insightPolicy *insightsv1alpha1.InsightPolicy, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightPolicy, error) {
	result := &insightsv1alpha1.InsightPolicy{}
	err := c.client.Put().
		Resource("insightpolicies").
		Name(insightPolicy.Name).
		VersionedParams(&opts, parameterCodec).
		Body(insightPolicy).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightPolicies) UpdateStatus(ctx context.Context, insightPolicy *insightsv1alpha1.InsightPolicy, opts metav1.UpdateOptions) (*insightsv1alpha1.InsightPolicy, error) {
	result := &insightsv1alpha1.InsightPolicy{}
	err := c.client.Put().
		Resource("insightpolicies").
		Name(insightPolicy.Name).
		SubResource("status").
		VersionedParams(&opts, parameterCodec).
		Body(insightPolicy).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insightPolicies) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("insightpolicies").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *insightPolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*insightsv1alpha1.InsightPolicy, error) {
	result := &insightsv1alpha1.InsightPolicy{}
	err := c.client.Patch(pt).
		Resource("insightpolicies").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, parameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}
