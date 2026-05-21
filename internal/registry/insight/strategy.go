package insight

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// insightStrategy implements behavior for Insight resources
type insightStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default strategy for Insight resources
var Strategy = insightStrategy{
	ObjectTyper:   nil, // Will be set from scheme
	NameGenerator: names.SimpleNameGenerator,
}

// NamespaceScoped returns true because Insights are namespaced
func (insightStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set on create
func (insightStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	insight := obj.(*insightsv1alpha1.Insight)

	// Initialize status
	if insight.Status.State == "" {
		insight.Status.State = insightsv1alpha1.InsightStateActive
	}

	// Add standard labels for querying
	if insight.Labels == nil {
		insight.Labels = make(map[string]string)
	}
	insight.Labels["insights.miloapis.com/state"] = string(insight.Status.State)
	insight.Labels["insights.miloapis.com/severity"] = string(insight.Spec.Severity)
	insight.Labels["insights.miloapis.com/category"] = insight.Spec.Category
}

// PrepareForUpdate clears fields that are not allowed to be set on update
func (insightStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newInsight := obj.(*insightsv1alpha1.Insight)
	oldInsight := old.(*insightsv1alpha1.Insight)

	// Preserve spec from old object (spec is immutable after creation)
	newInsight.Spec = oldInsight.Spec

	// Update labels to reflect current state
	if newInsight.Labels == nil {
		newInsight.Labels = make(map[string]string)
	}
	newInsight.Labels["insights.miloapis.com/state"] = string(newInsight.Status.State)
	newInsight.Labels["insights.miloapis.com/severity"] = string(newInsight.Spec.Severity)
	newInsight.Labels["insights.miloapis.com/category"] = newInsight.Spec.Category

	// Update owner label if set
	if newInsight.Status.Owner != nil {
		newInsight.Labels["insights.miloapis.com/owner-uid"] = newInsight.Status.Owner.UID
	} else {
		delete(newInsight.Labels, "insights.miloapis.com/owner-uid")
	}
}

// Validate validates a new Insight
func (insightStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	insight := obj.(*insightsv1alpha1.Insight)
	return validateInsight(insight)
}

// WarningsOnCreate returns warnings for the creation of the given object
func (insightStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// AllowCreateOnUpdate returns false because Insights are created via POST
func (insightStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate validates an update to an existing Insight
func (insightStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	newInsight := obj.(*insightsv1alpha1.Insight)
	oldInsight := old.(*insightsv1alpha1.Insight)
	return validateInsightUpdate(newInsight, oldInsight)
}

// WarningsOnUpdate returns warnings for the update of the given object
func (insightStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// AllowUnconditionalUpdate returns true to allow updates without resource version
func (insightStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// Canonicalize normalizes the object after validation
func (insightStrategy) Canonicalize(obj runtime.Object) {
}

// GetAttrs returns labels and fields of an Insight for filtering
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	insight, ok := obj.(*insightsv1alpha1.Insight)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not an Insight")
	}
	return insight.Labels, SelectableFields(insight), nil
}

// SelectableFields returns the fields that can be used in field selectors
func SelectableFields(insight *insightsv1alpha1.Insight) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&insight.ObjectMeta, true)
	specificFieldsSet := fields.Set{
		"spec.severity":       string(insight.Spec.Severity),
		"spec.category":       insight.Spec.Category,
		"status.state":        string(insight.Status.State),
		"spec.targetRef.kind": insight.Spec.TargetRef.Kind,
		"spec.targetRef.name": insight.Spec.TargetRef.Name,
		"spec.source.type":    string(insight.Spec.Source.Type),
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, specificFieldsSet)
}

// MatchInsight returns a generic matcher for Insight
func MatchInsight(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

func validateInsight(insight *insightsv1alpha1.Insight) field.ErrorList {
	allErrs := field.ErrorList{}

	// Validate required fields
	if insight.Spec.TargetRef.APIVersion == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "targetRef", "apiVersion"), ""))
	}
	if insight.Spec.TargetRef.Kind == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "targetRef", "kind"), ""))
	}
	if insight.Spec.TargetRef.Name == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "targetRef", "name"), ""))
	}
	if insight.Spec.Message == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "message"), ""))
	}
	if insight.Spec.Category == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "category"), ""))
	}

	return allErrs
}

func validateInsightUpdate(newInsight, oldInsight *insightsv1alpha1.Insight) field.ErrorList {
	allErrs := field.ErrorList{}

	// Validate state transitions
	allErrs = append(allErrs, validateStateTransition(oldInsight.Status.State, newInsight.Status.State)...)

	return allErrs
}

func validateStateTransition(oldState, newState insightsv1alpha1.InsightState) field.ErrorList {
	allErrs := field.ErrorList{}

	// Define valid transitions
	validTransitions := map[insightsv1alpha1.InsightState][]insightsv1alpha1.InsightState{
		"": {
			insightsv1alpha1.InsightStateActive,
		},
		insightsv1alpha1.InsightStateActive: {
			insightsv1alpha1.InsightStateAcknowledged,
			insightsv1alpha1.InsightStateSnoozed,
			insightsv1alpha1.InsightStateResolved,
		},
		insightsv1alpha1.InsightStateAcknowledged: {
			insightsv1alpha1.InsightStateActive,
			insightsv1alpha1.InsightStateSnoozed,
			insightsv1alpha1.InsightStateResolved,
		},
		insightsv1alpha1.InsightStateSnoozed: {
			insightsv1alpha1.InsightStateActive,
			insightsv1alpha1.InsightStateAcknowledged,
			insightsv1alpha1.InsightStateResolved,
		},
		insightsv1alpha1.InsightStateResolved: {
			// Resolved is terminal for user actions, but controller can clean up
		},
	}

	if oldState == newState {
		return allErrs
	}

	allowed, exists := validTransitions[oldState]
	if !exists {
		allErrs = append(allErrs, field.Invalid(
			field.NewPath("status", "state"),
			newState,
			fmt.Sprintf("unknown current state: %s", oldState),
		))
		return allErrs
	}

	for _, valid := range allowed {
		if newState == valid {
			return allErrs
		}
	}

	allErrs = append(allErrs, field.Invalid(
		field.NewPath("status", "state"),
		newState,
		fmt.Sprintf("invalid state transition from %s to %s", oldState, newState),
	))

	return allErrs
}
