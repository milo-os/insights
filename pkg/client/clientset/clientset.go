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

// Package clientset contains typed clients for the insights API server.
package clientset

import (
	"fmt"
	"net/http"

	"k8s.io/client-go/rest"

	insightsv1alpha1client "github.com/datum-cloud/insights/pkg/client/clientset/typed/insights/v1alpha1"
)

// Interface provides access to all the insights API group clients.
type Interface interface {
	InsightsV1alpha1() insightsv1alpha1client.InsightsV1alpha1Interface
}

// Clientset contains the clients for groups.
type Clientset struct {
	insightsV1alpha1 *insightsv1alpha1client.InsightsV1alpha1Client
}

// InsightsV1alpha1 retrieves the InsightsV1alpha1Client.
func (c *Clientset) InsightsV1alpha1() insightsv1alpha1client.InsightsV1alpha1Interface {
	return c.insightsV1alpha1
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c

	var cs Clientset
	var err error
	cs.insightsV1alpha1, err = insightsv1alpha1client.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigAndClient creates a new Clientset for the given config and http client.
func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*Clientset, error) {
	configShallowCopy := *c

	var cs Clientset
	var err error
	cs.insightsV1alpha1, err = insightsv1alpha1client.NewForConfigAndClient(&configShallowCopy, httpClient)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	cs, err := NewForConfig(c)
	if err != nil {
		panic(fmt.Sprintf("insights clientset: failed to create client: %v", err))
	}
	return cs
}
