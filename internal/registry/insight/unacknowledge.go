package insight

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// UnacknowledgeREST implements the /unacknowledge subresource
type UnacknowledgeREST struct {
	store *registry.Store
}

var (
	_ rest.Storage              = &UnacknowledgeREST{}
	_ rest.Scoper               = &UnacknowledgeREST{}
	_ rest.SingularNameProvider = &UnacknowledgeREST{}
	_ rest.NamedCreater         = &UnacknowledgeREST{}
)

// NewUnacknowledgeREST creates a new UnacknowledgeREST
func NewUnacknowledgeREST(store *registry.Store) *UnacknowledgeREST {
	return &UnacknowledgeREST{store: store}
}

// New returns an empty object
func (r *UnacknowledgeREST) New() runtime.Object {
	return &insightsv1alpha1.UnacknowledgeOptions{}
}

// Destroy cleans up resources on shutdown
func (r *UnacknowledgeREST) Destroy() {}

// NamespaceScoped returns true because Insights are namespaced
func (r *UnacknowledgeREST) NamespaceScoped() bool {
	return true
}

// GetSingularName returns the singular name of the resource
func (r *UnacknowledgeREST) GetSingularName() string {
	return "unacknowledge"
}

// Create handles POST /insights/{name}/unacknowledge
func (r *UnacknowledgeREST) Create(
	ctx context.Context,
	name string,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	_, ok := obj.(*insightsv1alpha1.UnacknowledgeOptions)
	if !ok {
		return nil, fmt.Errorf("expected UnacknowledgeOptions, got %T", obj)
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
		return nil, fmt.Errorf("cannot unacknowledge resolved insight")
	}
	if insight.Status.State == insightsv1alpha1.InsightStateActive {
		return nil, fmt.Errorf("insight is not acknowledged")
	}

	// Verify ownership - only owner can unacknowledge
	actor := actorFromUserInfo(userInfo)
	if insight.Status.Owner != nil && insight.Status.Owner.UID != actor.UID {
		return nil, fmt.Errorf("only the owner can unacknowledge an insight")
	}

	// Update insight
	insight.Status.State = insightsv1alpha1.InsightStateActive
	insight.Status.Owner = nil
	insight.Status.Acknowledgement = nil
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
