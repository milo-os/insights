package install

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	insightsv1alpha1 "github.com/datum-cloud/insights/pkg/apis/insights/v1alpha1"
)

func TestInstall(t *testing.T) {
	scheme := runtime.NewScheme()
	Install(scheme)

	// Test that v1alpha1 types are registered
	v1alpha1GV := schema.GroupVersion{Group: insightsv1alpha1.GroupName, Version: "v1alpha1"}

	types := []runtime.Object{
		&insightsv1alpha1.Insight{},
		&insightsv1alpha1.InsightList{},
		&insightsv1alpha1.InsightPolicy{},
		&insightsv1alpha1.InsightPolicyList{},
		&insightsv1alpha1.InsightMuteRule{},
		&insightsv1alpha1.InsightMuteRuleList{},
	}

	for _, obj := range types {
		gvks, _, err := scheme.ObjectKinds(obj)
		if err != nil {
			t.Errorf("Error getting kinds for %T: %v", obj, err)
			continue
		}
		if len(gvks) == 0 {
			t.Errorf("No kinds registered for %T", obj)
			continue
		}

		found := false
		for _, gvk := range gvks {
			if gvk.GroupVersion() == v1alpha1GV {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("v1alpha1 version not registered for %T, got: %v", obj, gvks)
		}
	}

	// Test that internal version is also registered
	internalGV := schema.GroupVersion{Group: insightsv1alpha1.GroupName, Version: runtime.APIVersionInternal}

	// Check if InsightPolicy has an internal version registration
	obj := &insightsv1alpha1.InsightPolicy{}
	gvks, _, err := scheme.ObjectKinds(obj)
	if err != nil {
		t.Fatalf("Error getting kinds for InsightPolicy: %v", err)
	}

	foundInternal := false
	for _, gvk := range gvks {
		if gvk.GroupVersion() == internalGV {
			foundInternal = true
			break
		}
	}
	if !foundInternal {
		t.Errorf("Internal version not registered for InsightPolicy, got: %v", gvks)
	}

	// Test version priority
	versions := scheme.PrioritizedVersionsForGroup(insightsv1alpha1.GroupName)
	if len(versions) == 0 {
		t.Fatal("No versions registered for group")
	}

	// v1alpha1 should be the preferred version
	if versions[0].Version != "v1alpha1" {
		t.Errorf("Expected v1alpha1 to be preferred version, got: %v", versions[0])
	}
}
