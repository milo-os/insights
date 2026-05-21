package insight

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// Storage contains the REST storage for Insight resources and subresources
type Storage struct {
	Insight       *REST
	Status        *StatusREST
	Acknowledge   *AcknowledgeREST
	Unacknowledge *UnacknowledgeREST
	Snooze        *SnoozeREST
	Resolve       *ResolveREST
	Assign        *AssignREST
}

// REST implements a RESTStorage for Insight
type REST struct {
	*genericregistry.Store
}

// NewStorage creates a new Storage instance
func NewStorage(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*Storage, error) {
	strategy := Strategy
	strategy.ObjectTyper = scheme

	store := &genericregistry.Store{
		NewFunc:                   func() runtime.Object { return &insightsv1alpha1.Insight{} },
		NewListFunc:               func() runtime.Object { return &insightsv1alpha1.InsightList{} },
		DefaultQualifiedResource:  insightsv1alpha1.Resource("insights"),
		SingularQualifiedResource: insightsv1alpha1.Resource("insight"),

		CreateStrategy:      strategy,
		UpdateStrategy:      strategy,
		DeleteStrategy:      strategy,
		ReturnDeletedObject: true,

		TableConvertor: rest.NewDefaultTableConvertor(insightsv1alpha1.Resource("insights")),
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	insightREST := &REST{store}

	return &Storage{
		Insight:       insightREST,
		Status:        NewStatusREST(store),
		Acknowledge:   NewAcknowledgeREST(store),
		Unacknowledge: NewUnacknowledgeREST(store),
		Snooze:        NewSnoozeREST(store),
		Resolve:       NewResolveREST(store),
		Assign:        NewAssignREST(store),
	}, nil
}

// ShortNames returns the short names for the resource
func (r *REST) ShortNames() []string {
	return []string{"ins"}
}

// Categories returns the categories for the resource
func (r *REST) Categories() []string {
	return []string{"insights"}
}

// Ensure REST implements required interfaces.
var (
	_ rest.Storage              = &REST{}
	_ rest.Getter               = &REST{}
	_ rest.Lister               = &REST{}
	_ rest.Creater              = &REST{} //nolint:misspell // rest.Creater is a k8s.io type name
	_ rest.Updater              = &REST{}
	_ rest.GracefulDeleter      = &REST{}
	_ rest.Watcher              = &REST{}
	_ rest.ShortNamesProvider   = &REST{}
	_ rest.CategoriesProvider   = &REST{}
	_ rest.SingularNameProvider = &REST{}
)
