package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupName is the group name used in this package
const GroupName = "insights.miloapis.com"

// SchemeGroupVersion is the group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha1"}

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	// localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
	localSchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// SchemeBuilder is the scheme builder with scheme init functions to run for this API package
	SchemeBuilder = localSchemeBuilder
	// AddToScheme is a common registration function for mapping types to a scheme
	AddToScheme = localSchemeBuilder.AddToScheme
)

// addKnownTypes adds our types to the API scheme by registering
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		// Core resources
		&Insight{},
		&InsightList{},
		&InsightPolicy{},
		&InsightPolicyList{},
		&InsightMuteRule{},
		&InsightMuteRuleList{},

		// Action request types (for subresources)
		&AcknowledgeOptions{},
		&SnoozeOptions{},
		&ResolveOptions{},
		&AssignOptions{},
		&UnacknowledgeOptions{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
