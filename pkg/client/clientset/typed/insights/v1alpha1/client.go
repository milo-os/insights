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
	"net/http"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"

	"github.com/datum-cloud/insights/pkg/apis/insights/install"
	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// InsightsV1alpha1Interface provides access to insights v1alpha1 resources.
type InsightsV1alpha1Interface interface {
	RESTClient() rest.Interface
	InsightsGetter
	InsightPoliciesGetter
	InsightMuteRulesGetter
}

var (
	scheme         = runtime.NewScheme()
	codecs         serializer.CodecFactory
	parameterCodec runtime.ParameterCodec
)

func init() {
	install.Install(scheme)
	codecs = serializer.NewCodecFactory(scheme)
	parameterCodec = runtime.NewParameterCodec(scheme)
}

// InsightsV1alpha1Client is used to interact with features provided by the
// insights.miloapis.com group.
type InsightsV1alpha1Client struct {
	restClient rest.Interface
}

func (c *InsightsV1alpha1Client) Insights(namespace string) InsightInterface {
	return newInsights(c, namespace)
}

func (c *InsightsV1alpha1Client) InsightPolicies() InsightPolicyInterface {
	return newInsightPolicies(c)
}

func (c *InsightsV1alpha1Client) InsightMuteRules(namespace string) InsightMuteRuleInterface {
	return newInsightMuteRules(c, namespace)
}

func (c *InsightsV1alpha1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}

// NewForConfig creates a new InsightsV1alpha1Client for the given config.
func NewForConfig(c *rest.Config) (*InsightsV1alpha1Client, error) {
	config := *c
	setConfigDefaults(&config)
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new InsightsV1alpha1Client for the given config and http client.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*InsightsV1alpha1Client, error) {
	config := *c
	setConfigDefaults(&config)
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &InsightsV1alpha1Client{client}, nil
}

// setConfigDefaults fills in default values for a REST config targeting the insights API group.
func setConfigDefaults(config *rest.Config) {
	gv := insightsv1alpha1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
}

// GroupVersion returns the schema.GroupVersion for the client.
func GroupVersion() schema.GroupVersion {
	return insightsv1alpha1.SchemeGroupVersion
}
