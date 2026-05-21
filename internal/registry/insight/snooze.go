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

// SnoozeREST implements the /snooze subresource
type SnoozeREST struct {
	store *registry.Store
}

var (
	_ rest.Storage              = &SnoozeREST{}
	_ rest.Scoper               = &SnoozeREST{}
	_ rest.SingularNameProvider = &SnoozeREST{}
	_ rest.NamedCreater         = &SnoozeREST{}
)

// NewSnoozeREST creates a new SnoozeREST
func NewSnoozeREST(store *registry.Store) *SnoozeREST {
	return &SnoozeREST{store: store}
}

// New returns an empty object
func (r *SnoozeREST) New() runtime.Object {
	return &insightsv1alpha1.SnoozeOptions{}
}

// Destroy cleans up resources on shutdown
func (r *SnoozeREST) Destroy() {}

// NamespaceScoped returns true because Insights are namespaced
func (r *SnoozeREST) NamespaceScoped() bool {
	return true
}

// GetSingularName returns the singular name of the resource
func (r *SnoozeREST) GetSingularName() string {
	return "snooze"
}

// Create handles POST /insights/{name}/snooze
func (r *SnoozeREST) Create(
	ctx context.Context,
	name string,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	opts, ok := obj.(*insightsv1alpha1.SnoozeOptions)
	if !ok {
		return nil, fmt.Errorf("expected SnoozeOptions, got %T", obj)
	}

	userInfo, ok := request.UserFrom(ctx)
	if !ok {
		return nil, fmt.Errorf("no user info in context")
	}

	// Calculate snooze until time
	var until time.Time
	if opts.Until != nil {
		until = opts.Until.Time
		if until.Before(time.Now()) {
			return nil, fmt.Errorf("snooze until time must be in the future")
		}
	} else if opts.Duration != "" {
		duration, err := time.ParseDuration(opts.Duration)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format: %w (use format like '1h', '4h', '24h', '168h')", err)
		}
		if duration <= 0 {
			return nil, fmt.Errorf("duration must be positive")
		}
		until = time.Now().Add(duration)
	} else {
		return nil, fmt.Errorf("either duration or until must be specified")
	}

	// Get current insight
	existing, err := r.store.Get(ctx, name, &metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	insight := existing.(*insightsv1alpha1.Insight)

	// Validate state
	if insight.Status.State == insightsv1alpha1.InsightStateResolved {
		return nil, fmt.Errorf("cannot snooze resolved insight")
	}

	actor := actorFromUserInfo(userInfo)
	now := metav1.NewTime(time.Now())

	// Update insight
	insight.Status.State = insightsv1alpha1.InsightStateSnoozed
	insight.Status.Owner = &actor
	insight.Status.Snooze = &insightsv1alpha1.SnoozeInfo{
		By:    actor,
		At:    now,
		Until: metav1.NewTime(until),
	}

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
