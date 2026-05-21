package install

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// Install registers the API group and adds types to a scheme
func Install(scheme *runtime.Scheme) {
	// Register v1alpha1 types
	utilruntime.Must(insightsv1alpha1.AddToScheme(scheme))

	// Set version priority (required for aggregated API servers)
	utilruntime.Must(scheme.SetVersionPriority(insightsv1alpha1.SchemeGroupVersion))

	// Register v1alpha1 as the internal version (hub version)
	// This allows the scheme to use v1alpha1 for both internal and external representations
	scheme.AddKnownTypes(schema.GroupVersion{
		Group:   insightsv1alpha1.GroupName,
		Version: runtime.APIVersionInternal,
	},
		&insightsv1alpha1.Insight{},
		&insightsv1alpha1.InsightList{},
		&insightsv1alpha1.InsightPolicy{},
		&insightsv1alpha1.InsightPolicyList{},
		&insightsv1alpha1.InsightMuteRule{},
		&insightsv1alpha1.InsightMuteRuleList{},
		&insightsv1alpha1.AcknowledgeOptions{},
		&insightsv1alpha1.UnacknowledgeOptions{},
		&insightsv1alpha1.SnoozeOptions{},
		&insightsv1alpha1.ResolveOptions{},
		&insightsv1alpha1.AssignOptions{},
	)
}
