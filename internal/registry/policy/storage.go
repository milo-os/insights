package policy

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	genericregistry "k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	insightcel "github.com/datum-cloud/insights/internal/cel"
	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// Storage contains the REST storage for InsightPolicy resources
type Storage struct {
	Policy *REST
	Status *StatusREST
}

// REST implements a RESTStorage for InsightPolicy
type REST struct {
	*genericregistry.Store
}

// policyStrategy implements behavior for InsightPolicy resources
type policyStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = policyStrategy{
	ObjectTyper:   nil,
	NameGenerator: names.SimpleNameGenerator,
}

// NewStorage creates a new Storage instance
func NewStorage(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*Storage, error) {
	strategy := Strategy
	strategy.ObjectTyper = scheme

	store := &genericregistry.Store{
		NewFunc:                   func() runtime.Object { return &insightsv1alpha1.InsightPolicy{} },
		NewListFunc:               func() runtime.Object { return &insightsv1alpha1.InsightPolicyList{} },
		DefaultQualifiedResource:  insightsv1alpha1.Resource("insightpolicies"),
		SingularQualifiedResource: insightsv1alpha1.Resource("insightpolicy"),

		CreateStrategy:      strategy,
		UpdateStrategy:      strategy,
		DeleteStrategy:      strategy,
		ReturnDeletedObject: true,

		TableConvertor: rest.NewDefaultTableConvertor(insightsv1alpha1.Resource("insightpolicies")),
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &Storage{
		Policy: &REST{store},
		Status: NewStatusREST(store),
	}, nil
}

// ShortNames returns the short names for the resource
func (r *REST) ShortNames() []string {
	return []string{"ip", "inspol"}
}

// Categories returns the categories for the resource
func (r *REST) Categories() []string {
	return []string{"insights"}
}

// NamespaceScoped returns false because InsightPolicies are cluster-scoped
func (policyStrategy) NamespaceScoped() bool {
	return false
}

func (policyStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	// No-op
}

func (policyStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newPolicy := obj.(*insightsv1alpha1.InsightPolicy)
	oldPolicy := old.(*insightsv1alpha1.InsightPolicy)

	// Preserve status
	newPolicy.Status = oldPolicy.Status
}

func (policyStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	policy := obj.(*insightsv1alpha1.InsightPolicy)
	return validatePolicy(policy)
}

func (policyStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (policyStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (policyStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validatePolicy(obj.(*insightsv1alpha1.InsightPolicy))
}

func (policyStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

func (policyStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (policyStrategy) Canonicalize(obj runtime.Object) {
}

func validatePolicy(policy *insightsv1alpha1.InsightPolicy) field.ErrorList {
	allErrs := field.ErrorList{}

	if policy.Spec.Selector.APIVersion == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "selector", "apiVersion"), ""))
	}
	if policy.Spec.Selector.Kind == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "selector", "kind"), ""))
	}

	// Validate matchExpression CEL if present
	if policy.Spec.Selector.MatchExpression != "" {
		if _, err := insightcel.CompileCondition(policy.Spec.Selector.MatchExpression); err != nil {
			allErrs = append(allErrs, field.Invalid(
				field.NewPath("spec", "selector", "matchExpression"),
				policy.Spec.Selector.MatchExpression,
				fmt.Sprintf("invalid matchExpression: %v", err),
			))
		}
	}

	if len(policy.Spec.Rules) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "rules"), "at least one rule is required"))
	}

	for i, rule := range policy.Spec.Rules {
		rulePath := field.NewPath("spec", "rules").Index(i)
		if rule.Name == "" {
			allErrs = append(allErrs, field.Required(rulePath.Child("name"), ""))
		}

		// Validate condition CEL expression
		if rule.Condition == "" {
			allErrs = append(allErrs, field.Required(rulePath.Child("condition"), ""))
		} else {
			if _, err := insightcel.CompileCondition(rule.Condition); err != nil {
				allErrs = append(allErrs, field.Invalid(
					rulePath.Child("condition"),
					rule.Condition,
					fmt.Sprintf("rule %q (index %d): invalid condition: %v", rule.Name, i, err),
				))
			}
		}

		// Validate messageTemplate
		if rule.MessageTemplate == "" {
			allErrs = append(allErrs, field.Required(rulePath.Child("messageTemplate"), ""))
		} else {
			if err := insightcel.ValidateTemplate(rule.MessageTemplate); err != nil {
				allErrs = append(allErrs, field.Invalid(
					rulePath.Child("messageTemplate"),
					rule.MessageTemplate,
					fmt.Sprintf("rule %q (index %d): invalid messageTemplate: %v", rule.Name, i, err),
				))
			}
		}

		// Validate descriptionTemplate if present
		if rule.DescriptionTemplate != "" {
			if err := insightcel.ValidateTemplate(rule.DescriptionTemplate); err != nil {
				allErrs = append(allErrs, field.Invalid(
					rulePath.Child("descriptionTemplate"),
					rule.DescriptionTemplate,
					fmt.Sprintf("rule %q (index %d): invalid descriptionTemplate: %v", rule.Name, i, err),
				))
			}
		}

		if rule.Category == "" {
			allErrs = append(allErrs, field.Required(rulePath.Child("category"), ""))
		}
	}

	return allErrs
}

// GetAttrs returns labels and fields of an InsightPolicy for filtering
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	policy, ok := obj.(*insightsv1alpha1.InsightPolicy)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not an InsightPolicy")
	}
	return policy.Labels, SelectableFields(policy), nil
}

// SelectableFields returns the fields that can be used in field selectors
func SelectableFields(policy *insightsv1alpha1.InsightPolicy) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&policy.ObjectMeta, false)
	specificFieldsSet := fields.Set{
		"spec.selector.kind":       policy.Spec.Selector.Kind,
		"spec.selector.apiVersion": policy.Spec.Selector.APIVersion,
		"spec.suspended":           fmt.Sprintf("%t", policy.Spec.Suspended),
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, specificFieldsSet)
}

// MatchPolicy returns a generic matcher for InsightPolicy
func MatchPolicy(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// StatusREST implements the /status subresource for InsightPolicy
type StatusREST struct {
	store *genericregistry.Store
}

func NewStatusREST(store *genericregistry.Store) *StatusREST {
	return &StatusREST{store: store}
}

func (r *StatusREST) New() runtime.Object {
	return &insightsv1alpha1.InsightPolicy{}
}

func (r *StatusREST) Destroy() {}

func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

func (r *StatusREST) Update(
	ctx context.Context,
	name string,
	objInfo rest.UpdatedObjectInfo,
	createValidation rest.ValidateObjectFunc,
	updateValidation rest.ValidateObjectUpdateFunc,
	forceAllowCreate bool,
	options *metav1.UpdateOptions,
) (runtime.Object, bool, error) {
	return r.store.Update(ctx, name, objInfo, createValidation, updateValidation, forceAllowCreate, options)
}

var (
	_ rest.Storage = &StatusREST{}
	_ rest.Getter  = &StatusREST{}
	_ rest.Updater = &StatusREST{}
)
