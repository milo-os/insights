package app

import (
	"context"
	"fmt"
	"net"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/apiserver/pkg/server/options"
	utilversion "k8s.io/component-base/version"

	"github.com/datum-cloud/insights/internal/apiserver"
	generatedopenapi "github.com/datum-cloud/insights/pkg/generated/openapi"
)

const defaultEtcdPathPrefix = "/registry/insights.miloapis.com"

// Version is set at build time via -ldflags
var Version = "0.0.0-dev"

// Options contains the options for running the insights apiserver
type Options struct {
	RecommendedOptions *options.RecommendedOptions
}

// NewOptions creates a new Options with default values
func NewOptions() *Options {
	o := &Options{
		RecommendedOptions: options.NewRecommendedOptions(
			defaultEtcdPathPrefix,
			apiserver.Codecs.LegacyCodec(apiserver.Scheme.PrioritizedVersionsAllGroups()...),
		),
	}
	return o
}

// NewServerCommand creates a new cobra command for running the insights apiserver
func NewServerCommand() *cobra.Command {
	o := NewOptions()

	cmd := &cobra.Command{
		Use:   "insights-apiserver",
		Short: "Launch the insights API server",
		Long:  "Launch the insights API server providing custom resources for managing insights.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			return o.Run(cmd.Context())
		},
	}

	flags := cmd.Flags()
	o.RecommendedOptions.AddFlags(flags)

	return cmd
}

// Complete fills in any fields not set that are required to have valid data
func (o *Options) Complete() error {
	return nil
}

// Validate validates the options
func (o *Options) Validate() error {
	var errs []error
	errs = append(errs, o.RecommendedOptions.Validate()...)
	return utilerrors.NewAggregate(errs)
}

// Config returns the apiserver config
func (o *Options) Config() (*apiserver.Config, error) {
	// Generate self-signed certs with SANs for both localhost and the service DNS names
	// This allows the apiserver to work in both local and in-cluster scenarios
	alternateNames := []string{
		"localhost",
		"insights-apiserver.insights-system.svc",
		"insights-apiserver.insights-system.svc.cluster.local",
	}
	alternateIPs := []net.IP{net.ParseIP("127.0.0.1")}

	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts(
		"localhost",
		alternateNames,
		alternateIPs,
	); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %w", err)
	}

	serverConfig := server.NewRecommendedConfig(apiserver.Codecs)
	serverConfig.EffectiveVersion = utilversion.NewEffectiveVersion(Version)
	serverConfig.OpenAPIV3Config = server.DefaultOpenAPIV3Config(
		generatedopenapi.GetOpenAPIDefinitions,
		openapi.NewDefinitionNamer(apiserver.Scheme),
	)
	serverConfig.OpenAPIV3Config.Info.Title = "Insights API"
	serverConfig.OpenAPIV3Config.Info.Version = Version

	if err := o.RecommendedOptions.ApplyTo(serverConfig); err != nil {
		return nil, err
	}

	config := &apiserver.Config{
		GenericConfig: serverConfig,
		ExtraConfig:   apiserver.ExtraConfig{},
	}

	return config, nil
}

// Run starts the insights apiserver
func (o *Options) Run(ctx context.Context) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	srv, err := config.Complete().New(ctx)
	if err != nil {
		return err
	}

	return srv.GenericAPIServer.PrepareRun().RunWithContext(ctx)
}
