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

// InsightsGetter has a method to return an InsightInterface.
type InsightsGetter interface {
	Insights(namespace string) InsightInterface
}

// InsightInterface has methods to work with Insight resources.
type InsightInterface interface {
	Create(ctx context.Context, insight *insightsv1alpha1.Insight, opts metav1.CreateOptions) (*insightsv1alpha1.Insight, error)
	Update(ctx context.Context, insight *insightsv1alpha1.Insight, opts metav1.UpdateOptions) (*insightsv1alpha1.Insight, error)
	UpdateStatus(ctx context.Context, insight *insightsv1alpha1.Insight, opts metav1.UpdateOptions) (*insightsv1alpha1.Insight, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*insightsv1alpha1.Insight, error)
	List(ctx context.Context, opts metav1.ListOptions) (*insightsv1alpha1.InsightList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*insightsv1alpha1.Insight, error)
}

// insights implements InsightInterface.
type insights struct {
	client rest.Interface
	ns     string
}

// newInsights returns an insights client.
func newInsights(c *InsightsV1alpha1Client, namespace string) *insights {
	return &insights{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

func (c *insights) Get(ctx context.Context, name string, opts metav1.GetOptions) (*insightsv1alpha1.Insight, error) {
	result := &insightsv1alpha1.Insight{}
	req := c.client.Get().Resource("insights").Name(name).VersionedParams(&opts, parameterCodec)
	if c.ns != "" {
		req = c.client.Get().Resource("insights").Namespace(c.ns).Name(name).VersionedParams(&opts, parameterCodec)
	}
	err := req.Do(ctx).Into(result)
	return result, err
}

func (c *insights) List(ctx context.Context, opts metav1.ListOptions) (*insightsv1alpha1.InsightList, error) {
	result := &insightsv1alpha1.InsightList{}
	var req *rest.Request
	if c.ns != "" {
		req = c.client.Get().Resource("insights").Namespace(c.ns).VersionedParams(&opts, parameterCodec)
	} else {
		req = c.client.Get().Resource("insights").VersionedParams(&opts, parameterCodec)
	}
	err := req.Do(ctx).Into(result)
	return result, err
}

func (c *insights) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	var req *rest.Request
	if c.ns != "" {
		req = c.client.Get().Resource("insights").Namespace(c.ns).VersionedParams(&opts, parameterCodec)
	} else {
		req = c.client.Get().Resource("insights").VersionedParams(&opts, parameterCodec)
	}
	return req.Watch(ctx)
}

func (c *insights) Create(ctx context.Context, insight *insightsv1alpha1.Insight, opts metav1.CreateOptions) (*insightsv1alpha1.Insight, error) {
	result := &insightsv1alpha1.Insight{}
	err := c.client.Post().
		Namespace(c.ns).
		Resource("insights").
		VersionedParams(&opts, parameterCodec).
		Body(insight).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insights) Update(ctx context.Context, insight *insightsv1alpha1.Insight, opts metav1.UpdateOptions) (*insightsv1alpha1.Insight, error) {
	result := &insightsv1alpha1.Insight{}
	err := c.client.Put().
		Namespace(c.ns).
		Resource("insights").
		Name(insight.Name).
		VersionedParams(&opts, parameterCodec).
		Body(insight).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insights) UpdateStatus(ctx context.Context, insight *insightsv1alpha1.Insight, opts metav1.UpdateOptions) (*insightsv1alpha1.Insight, error) {
	result := &insightsv1alpha1.Insight{}
	err := c.client.Put().
		Namespace(c.ns).
		Resource("insights").
		Name(insight.Name).
		SubResource("status").
		VersionedParams(&opts, parameterCodec).
		Body(insight).
		Do(ctx).
		Into(result)
	return result, err
}

func (c *insights) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("insights").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

func (c *insights) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*insightsv1alpha1.Insight, error) {
	result := &insightsv1alpha1.Insight{}
	err := c.client.Patch(pt).
		Namespace(c.ns).
		Resource("insights").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, parameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return result, err
}
