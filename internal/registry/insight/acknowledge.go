package insight

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// AcknowledgeREST implements the /acknowledge subresource
type AcknowledgeREST struct {
	store *registry.Store
}

var (
	_ rest.Storage              = &AcknowledgeREST{}
	_ rest.Scoper               = &AcknowledgeREST{}
	_ rest.SingularNameProvider = &AcknowledgeREST{}
	_ rest.NamedCreater         = &AcknowledgeREST{}
)

// NewAcknowledgeREST creates a new AcknowledgeREST
func NewAcknowledgeREST(store *registry.Store) *AcknowledgeREST {
	return &AcknowledgeREST{store: store}
}

// New returns an empty object
func (r *AcknowledgeREST) New() runtime.Object {
	return &insightsv1alpha1.AcknowledgeOptions{}
}

// Destroy cleans up resources on shutdown
func (r *AcknowledgeREST) Destroy() {}

// NamespaceScoped returns true because Insights are namespaced
func (r *AcknowledgeREST) NamespaceScoped() bool {
	return true
}

// GetSingularName returns the singular name of the resource
func (r *AcknowledgeREST) GetSingularName() string {
	return "acknowledge"
}

// Create handles POST /insights/{name}/acknowledge
func (r *AcknowledgeREST) Create(
	ctx context.Context,
	name string,
	obj runtime.Object,
	createValidation rest.ValidateObjectFunc,
	options *metav1.CreateOptions,
) (runtime.Object, error) {
	opts, ok := obj.(*insightsv1alpha1.AcknowledgeOptions)
	if !ok {
		return nil, fmt.Errorf("expected AcknowledgeOptions, got %T", obj)
	}

	// Get the authenticated user from context
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

	// Validate state transition
	if insight.Status.State == insightsv1alpha1.InsightStateResolved {
		return nil, fmt.Errorf("cannot acknowledge resolved insight")
	}
	if insight.Status.State == insightsv1alpha1.InsightStateAcknowledged {
		return nil, fmt.Errorf("insight is already acknowledged")
	}

	// Build actor from user info
	actor := actorFromUserInfo(userInfo)
	now := metav1.NewTime(time.Now())

	// Update insight
	insight.Status.State = insightsv1alpha1.InsightStateAcknowledged
	insight.Status.Owner = &actor
	insight.Status.Acknowledgement = &insightsv1alpha1.AcknowledgementInfo{
		By:   actor,
		At:   now,
		Note: opts.Note,
	}
	// Clear snooze if it was snoozed
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

// actorFromUserInfo converts Kubernetes user info to an ActorReference
func actorFromUserInfo(u user.Info) insightsv1alpha1.ActorReference {
	actorType := "user"
	if strings.HasPrefix(u.GetName(), "system:serviceaccount:") {
		actorType = "serviceaccount"
	}

	email := ""
	// Try to extract email from user info
	if extras := u.GetExtra(); extras != nil {
		if emails, ok := extras["email"]; ok && len(emails) > 0 {
			email = emails[0]
		}
	}
	// Fall back to name if it looks like an email
	if email == "" && strings.Contains(u.GetName(), "@") {
		email = u.GetName()
	}

	return insightsv1alpha1.ActorReference{
		Type:  actorType,
		Name:  u.GetName(),
		UID:   u.GetUID(),
		Email: email,
	}
}
