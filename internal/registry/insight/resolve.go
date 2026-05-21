package insight

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

const (
	// ResolvedRetentionDays is how long resolved insights are kept before deletion
	ResolvedRetentionDays = 30
)

// ResolveREST implements the /resolve subresource
type ResolveREST struct {
	store *registry.Store
}

var (
	_ rest.Storage              = &ResolveREST{}
	_ rest.Scoper               = &ResolveREST{}
	_ rest.SingularNameProvider = &ResolveREST{}
	_ rest.NamedCreater         = &ResolveREST{}
)

// NewResolveREST creates a new ResolveREST
func NewResolveREST(store *registry.Store) *ResolveREST {
	return &ResolveREST{store: store}
}

// New returns an empty object
func (r *ResolveREST) New() runtime.Object {
	return &insightsv1alpha1.ResolveOptions{}
}

// Destroy cleans up resources on shutdown
func (r *ResolveREST) Destroy() {}

// NamespaceScoped returns true because Insights are namespaced
func (r *ResolveREST) NamespaceScoped() bool {
	return true
}

// GetSingularName returns the singular name of the resource
func (r *ResolveREST) GetSingularName() string {
	return "resolve"
}

// Create handles POST /insights/{name}/resolve
func (r *ResolveREST) Create(
	ctx context.Context,
	name string,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	opts, ok := obj.(*insightsv1alpha1.ResolveOptions)
	if !ok {
		return nil, fmt.Errorf("expected ResolveOptions, got %T", obj)
	}

	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("no user info in context")
	}

	// Get current insight
	existing, err := r.store.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	insight := existing.(*insightsv1alpha1.Insight)

	// Validate state
	if insight.Status.State == insightsv1alpha1.InsightStateResolved {
		return nil, fmt.Errorf("insight is already resolved")
	}

	actor := actorFromUserInfo(userInfo)
	now := time.Now()

	// Update insight
	insight.Status.State = insightsv1alpha1.InsightStateResolved
	insight.Status.Resolution = &insightsv1alpha1.ResolutionInfo{
		By:   actor,
		At:   metav1.NewTime(now),
		Note: opts.Note,
	}

	// Set deletion time to 30 days from now
	deleteAt := now.AddDate(0, 0, ResolvedRetentionDays)
	insight.Status.DeleteAt = &metav1.Time{Time: deleteAt}

	// Clear snooze
	insight.Status.Snooze = nil

	// Persist update
	updated, _, err := r.store.Update(
		ctx,
		insight.Name,
		rest.DefaultUpdatedObjectInfo(insight),
		nil, nil, false,
		&metav1.UpdateOptions{},
	)
	if err != nil {
		return nil, err
	}

	return updated, nil
}
