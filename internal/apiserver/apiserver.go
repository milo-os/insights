package apiserver

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"

	"github.com/datum-cloud/insights/pkg/apis/insights/install"
	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"

	insightregistry "github.com/datum-cloud/insights/internal/registry/insight"
	muteruleregistry "github.com/datum-cloud/insights/internal/registry/muterule"
	policyregistry "github.com/datum-cloud/insights/internal/registry/policy"
)

var (
	// Scheme defines methods for serializing and deserializing API objects
	Scheme = runtime.NewScheme()
	// Codecs provides methods for retrieving codecs and serializers for specific versions and content types
	Codecs = serializer.NewCodecFactory(Scheme)
)

func init() {
	install.Install(Scheme)

	// Add unversioned types
	metav1.AddToGroupVersion(Scheme, schema.GroupVersion{Version: "v1"})

	unversioned := schema.GroupVersion{Group: "", Version: "v1"}
	Scheme.AddUnversionedTypes(unversioned,
		&metav1.Status{},
		&metav1.APIVersions{},
		&metav1.APIGroupList{},
		&metav1.APIGroup{},
		&metav1.APIResourceList{},
	)
}

// ExtraConfig holds custom apiserver config
type ExtraConfig struct {
	// Add any custom configuration here
	// For example: etcd client, event recorder, etc.
}

// Config defines the config for the apiserver
type Config struct {
	GenericConfig *genericapiserver.RecommendedConfig
	ExtraConfig   ExtraConfig
}

type completedConfig struct {
	GenericConfig genericapiserver.CompletedConfig
	ExtraConfig   *ExtraConfig
}

// CompletedConfig embeds a private pointer that cannot be instantiated outside of this package
type CompletedConfig struct {
	*completedConfig
}

// Complete fills in any fields not set that are required to have valid data
func (cfg *Config) Complete() CompletedConfig {
	c := completedConfig{
		cfg.GenericConfig.Complete(),
		&cfg.ExtraConfig,
	}

	return CompletedConfig{&c}
}

// Server contains state for a Kubernetes custom api server
type Server struct {
	GenericAPIServer *genericapiserver.GenericAPIServer
}

// New returns a new instance of Server from the given config
func (c completedConfig) New(ctx context.Context) (*Server, error) {
	genericServer, err := c.GenericConfig.New("insights-apiserver", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return nil, err
	}

	s := &Server{
		GenericAPIServer: genericServer,
	}

	// Create storage backends
	insightStorage, err := insightregistry.NewStorage(Scheme, c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, err
	}

	policyStorage, err := policyregistry.NewStorage(Scheme, c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, err
	}

	muteRuleStorage, err := muteruleregistry.NewStorage(Scheme, c.GenericConfig.RESTOptionsGetter)
	if err != nil {
		return nil, err
	}

	// Build API group
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(
		insightsv1alpha1.GroupName,
		Scheme,
		metav1.ParameterCodec,
		Codecs,
	)

	v1alpha1Storage := map[string]rest.Storage{
		// Main resources
		"insights":         insightStorage.Insight,
		"insightpolicies":  policyStorage.Policy,
		"insightmuterules": muteRuleStorage.MuteRule,

		// Insight subresources
		"insights/status":        insightStorage.Status,
		"insights/acknowledge":   insightStorage.Acknowledge,
		"insights/unacknowledge": insightStorage.Unacknowledge,
		"insights/snooze":        insightStorage.Snooze,
		"insights/resolve":       insightStorage.Resolve,
		"insights/assign":        insightStorage.Assign,

		// Policy subresources
		"insightpolicies/status": policyStorage.Status,

		// MuteRule subresources
		"insightmuterules/status": muteRuleStorage.Status,
	}

	apiGroupInfo.VersionedResourcesStorageMap["v1alpha1"] = v1alpha1Storage

	if err := s.GenericAPIServer.InstallAPIGroup(&apiGroupInfo); err != nil {
		return nil, err
	}

	return s, nil
}
