package apiserver

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

func TestSchemeRegistration(t *testing.T) {
	// Test that the scheme has all required types registered
	gvk := schema.GroupVersionKind{
		Group:   insightsv1alpha1.GroupName,
		Version: "v1alpha1",
		Kind:    "InsightPolicy",
	}

	obj, err := Scheme.New(gvk)
	if err != nil {
		t.Fatalf("Failed to create InsightPolicy from scheme: %v", err)
	}

	if _, ok := obj.(*insightsv1alpha1.InsightPolicy); !ok {
		t.Errorf("Expected *InsightPolicy, got %T", obj)
	}

	// Test internal version registration
	internalGVK := schema.GroupVersionKind{
		Group:   insightsv1alpha1.GroupName,
		Version: runtime.APIVersionInternal,
		Kind:    "InsightPolicy",
	}

	internalObj, err := Scheme.New(internalGVK)
	if err != nil {
		t.Fatalf("Failed to create InsightPolicy from internal version: %v", err)
	}

	if _, ok := internalObj.(*insightsv1alpha1.InsightPolicy); !ok {
		t.Errorf("Expected *InsightPolicy for internal version, got %T", internalObj)
	}
}

func TestInsightPolicyConversion(t *testing.T) {
	// Create a v1alpha1 InsightPolicy
	policy := &insightsv1alpha1.InsightPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-policy",
		},
		Spec: insightsv1alpha1.InsightPolicySpec{
			Selector: insightsv1alpha1.ResourceSelector{
				APIVersion: "v1",
				Kind:       "Pod",
			},
			Rules: []insightsv1alpha1.InsightRule{
				{
					Name:            "test-rule",
					Condition:       "true",
					Severity:        insightsv1alpha1.InsightSeverityInfo,
					Category:        "test",
					MessageTemplate: "Test message",
				},
			},
		},
	}

	// Test conversion to internal version
	internalGV := schema.GroupVersion{
		Group:   insightsv1alpha1.GroupName,
		Version: runtime.APIVersionInternal,
	}

	internalObj, err := Scheme.ConvertToVersion(policy, internalGV)
	if err != nil {
		t.Fatalf("Failed to convert to internal version: %v", err)
	}

	// Should be the same type since we're using v1alpha1 as the hub
	if _, ok := internalObj.(*insightsv1alpha1.InsightPolicy); !ok {
		t.Errorf("Expected *InsightPolicy after conversion, got %T", internalObj)
	}

	// Test conversion back to v1alpha1
	v1alpha1Obj, err := Scheme.ConvertToVersion(internalObj, insightsv1alpha1.SchemeGroupVersion)
	if err != nil {
		t.Fatalf("Failed to convert back to v1alpha1: %v", err)
	}

	convertedPolicy, ok := v1alpha1Obj.(*insightsv1alpha1.InsightPolicy)
	if !ok {
		t.Fatalf("Expected *InsightPolicy, got %T", v1alpha1Obj)
	}

	// Verify the data is preserved
	if convertedPolicy.Name != policy.Name {
		t.Errorf("Name mismatch: expected %s, got %s", policy.Name, convertedPolicy.Name)
	}

	if len(convertedPolicy.Spec.Rules) != len(policy.Spec.Rules) {
		t.Errorf("Rules count mismatch: expected %d, got %d", len(policy.Spec.Rules), len(convertedPolicy.Spec.Rules))
	}
}

func TestAllTypesRegistered(t *testing.T) {
	// Test all types can be created from the scheme
	types := []struct {
		gvk  schema.GroupVersionKind
		want runtime.Object
	}{
		{
			gvk: schema.GroupVersionKind{
				Group:   insightsv1alpha1.GroupName,
				Version: "v1alpha1",
				Kind:    "Insight",
			},
			want: &insightsv1alpha1.Insight{},
		},
		{
			gvk: schema.GroupVersionKind{
				Group:   insightsv1alpha1.GroupName,
				Version: "v1alpha1",
				Kind:    "InsightPolicy",
			},
			want: &insightsv1alpha1.InsightPolicy{},
		},
		{
			gvk: schema.GroupVersionKind{
				Group:   insightsv1alpha1.GroupName,
				Version: "v1alpha1",
				Kind:    "InsightMuteRule",
			},
			want: &insightsv1alpha1.InsightMuteRule{},
		},
	}

	for _, tt := range types {
		t.Run(tt.gvk.Kind, func(t *testing.T) {
			obj, err := Scheme.New(tt.gvk)
			if err != nil {
				t.Fatalf("Failed to create %s: %v", tt.gvk.Kind, err)
			}

			if reflect.TypeOf(obj) != reflect.TypeOf(tt.want) {
				t.Errorf("Type mismatch: expected %T, got %T", tt.want, obj)
			}
		})
	}
}
