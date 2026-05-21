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

// AssignREST implements the /assign subresource
type AssignREST struct {
	store *registry.Store
}

var (
	_ rest.Storage              = &AssignREST{}
	_ rest.Scoper               = &AssignREST{}
	_ rest.SingularNameProvider = &AssignREST{}
	_ rest.NamedCreater         = &AssignREST{}
)

// NewAssignREST creates a new AssignREST
func NewAssignREST(store *registry.Store) *AssignREST {
	return &AssignREST{store: store}
}

// New returns an empty object
func (r *AssignREST) New() runtime.Object {
	return &insightsv1alpha1.AssignOptions{}
}

// Destroy cleans up resources on shutdown
func (r *AssignREST) Destroy() {}

// NamespaceScoped returns true because Insights are namespaced
func (r *AssignREST) NamespaceScoped() bool {
	return true
}

// GetSingularName returns the singular name of the resource
func (r *AssignREST) GetSingularName() string {
	return "assign"
}

// Create handles POST /insights/{name}/assign
func (r *AssignREST) Create(
	ctx context.Context,
	name string,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	opts, ok := obj.(*insightsv1alpha1.AssignOptions)
	if !ok {
		return nil, fmt.Errorf("expected AssignOptions, got %T", obj)
	}

	// Validate assignee
	if opts.Assignee.Name == "" {
		return nil, fmt.Errorf("assignee name is required")
	}
	if opts.Assignee.Type == "" {
		return nil, fmt.Errorf("assignee type is required")
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
		return nil, fmt.Errorf("cannot assign resolved insight")
	}

	actor := actorFromUserInfo(userInfo)
	now := metav1.NewTime(time.Now())

	// Update insight
	insight.Status.Owner = &opts.Assignee
	insight.Status.Assignment = &insightsv1alpha1.AssignmentInfo{
		By:   actor,
		To:   opts.Assignee,
		At:   now,
		Note: opts.Note,
	}

	// If the insight was snoozed and is being reassigned, keep it snoozed
	// The new owner inherits the snooze state

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
