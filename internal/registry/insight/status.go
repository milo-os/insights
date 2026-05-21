package insight

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/generic/registry"
	"k8s.io/apiserver/pkg/registry/rest"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

// StatusREST implements the /status subresource
type StatusREST struct {
	store *registry.Store
}

var (
	_ rest.Storage         = &StatusREST{}
	_ rest.Getter          = &StatusREST{}
	_ rest.Updater         = &StatusREST{}
	_ rest.Patcher         = &StatusREST{}
	_ rest.StorageMetadata = &StatusREST{}
)

// NewStatusREST creates a new StatusREST
func NewStatusREST(store *registry.Store) *StatusREST {
	return &StatusREST{store: store}
}

// New returns an empty object
func (r *StatusREST) New() runtime.Object {
	return &insightsv1alpha1.Insight{}
}

// Destroy cleans up resources on shutdown
func (r *StatusREST) Destroy() {}

// Get retrieves the object
func (r *StatusREST) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	return r.store.Get(ctx, name, options)
}

// Update updates the status of an Insight
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

// ProducesMIMETypes returns the supported mime types
func (r *StatusREST) ProducesMIMETypes(verb string) []string {
	return nil
}

// ProducesObject returns the object type
func (r *StatusREST) ProducesObject(verb string) interface{} {
	return &insightsv1alpha1.Insight{}
}
