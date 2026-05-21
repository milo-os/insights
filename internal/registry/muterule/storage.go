package muterule

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

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// Storage contains the REST storage for InsightMuteRule resources
type Storage struct {
	MuteRule *REST
	Status   *StatusREST
}

// REST implements a RESTStorage for InsightMuteRule
type REST struct {
	*genericregistry.Store
}

// muteRuleStrategy implements behavior for InsightMuteRule resources
type muteRuleStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

var Strategy = muteRuleStrategy{
	ObjectTyper:   nil,
	NameGenerator: names.SimpleNameGenerator,
}

// NewStorage creates a new Storage instance
func NewStorage(scheme *runtime.Scheme, optsGetter generic.RESTOptionsGetter) (*Storage, error) {
	strategy := Strategy
	strategy.ObjectTyper = scheme

	store := &genericregistry.Store{
		NewFunc:                   func() runtime.Object { return &insightsv1alpha1.InsightMuteRule{} },
		NewListFunc:               func() runtime.Object { return &insightsv1alpha1.InsightMuteRuleList{} },
		DefaultQualifiedResource:  insightsv1alpha1.Resource("insightmuterules"),
		SingularQualifiedResource: insightsv1alpha1.Resource("insightmuterule"),

		CreateStrategy:      strategy,
		UpdateStrategy:      strategy,
		DeleteStrategy:      strategy,
		ReturnDeletedObject: true,

		TableConvertor: rest.NewDefaultTableConvertor(insightsv1alpha1.Resource("insightmuterules")),
	}

	options := &generic.StoreOptions{
		RESTOptions: optsGetter,
	}
	if err := store.CompleteWithOptions(options); err != nil {
		return nil, err
	}

	return &Storage{
		MuteRule: &REST{store},
		Status:   NewStatusREST(store),
	}, nil
}

// ShortNames returns the short names for the resource
func (r *REST) ShortNames() []string {
	return []string{"imr", "mute"}
}

// Categories returns the categories for the resource
func (r *REST) Categories() []string {
	return []string{"insights"}
}

// NamespaceScoped returns true because InsightMuteRules are namespaced
func (muteRuleStrategy) NamespaceScoped() bool {
	return true
}

func (muteRuleStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	// No-op
}

func (muteRuleStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newRule := obj.(*insightsv1alpha1.InsightMuteRule)
	oldRule := old.(*insightsv1alpha1.InsightMuteRule)

	// Preserve status
	newRule.Status = oldRule.Status
}

func (muteRuleStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	rule := obj.(*insightsv1alpha1.InsightMuteRule)
	return validateMuteRule(rule)
}

func (muteRuleStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

func (muteRuleStrategy) AllowCreateOnUpdate() bool {
	return false
}

func (muteRuleStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return validateMuteRule(obj.(*insightsv1alpha1.InsightMuteRule))
}

func (muteRuleStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

func (muteRuleStrategy) AllowUnconditionalUpdate() bool {
	return false
}

func (muteRuleStrategy) Canonicalize(obj runtime.Object) {
}

func validateMuteRule(rule *insightsv1alpha1.InsightMuteRule) field.ErrorList {
	allErrs := field.ErrorList{}

	// At least one match criteria must be specified
	match := rule.Spec.Match
	if match.PolicyRef == nil && match.Category == "" && match.TargetRef == nil && match.Severity == nil && match.LabelSelector == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("spec", "match"), "at least one match criteria must be specified"))
	}

	return allErrs
}

// GetAttrs returns labels and fields of an InsightMuteRule for filtering
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	rule, ok := obj.(*insightsv1alpha1.InsightMuteRule)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not an InsightMuteRule")
	}
	return rule.Labels, SelectableFields(rule), nil
}

// SelectableFields returns the fields that can be used in field selectors
func SelectableFields(rule *insightsv1alpha1.InsightMuteRule) fields.Set {
	objectMetaFieldsSet := generic.ObjectMetaFieldsSet(&rule.ObjectMeta, true)
	specificFieldsSet := fields.Set{
		"spec.match.category": rule.Spec.Match.Category,
	}
	if rule.Spec.Match.Severity != nil {
		specificFieldsSet["spec.match.severity"] = string(*rule.Spec.Match.Severity)
	}
	return generic.MergeFieldsSets(objectMetaFieldsSet, specificFieldsSet)
}

// MatchMuteRule returns a generic matcher for InsightMuteRule
func MatchMuteRule(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// StatusREST implements the /status subresource for InsightMuteRule
type StatusREST struct {
	store *genericregistry.Store
}

func NewStatusREST(store *genericregistry.Store) *StatusREST {
	return &StatusREST{store: store}
}

func (r *StatusREST) New() runtime.Object {
	return &insightsv1alpha1.InsightMuteRule{}
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
