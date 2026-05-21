package policy

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/datum-cloud/insights/pkg/apis/insights/install"
)

func TestStrategyObjectTyperInitialization(t *testing.T) {
	// Create a scheme
	scheme := runtime.NewScheme()
	install.Install(scheme)

	// Verify the default Strategy has nil ObjectTyper
	if Strategy.ObjectTyper != nil {
		t.Error("Expected default Strategy.ObjectTyper to be nil before initialization")
	}

	// Simulate what NewStorage does - create a copy and set the ObjectTyper
	strategy := Strategy
	strategy.ObjectTyper = scheme

	// Verify ObjectTyper is now set
	if strategy.ObjectTyper == nil {
		t.Fatal("Strategy ObjectTyper should not be nil after initialization")
	}

	// Verify it's the correct scheme
	if strategy.ObjectTyper != scheme {
		t.Error("Strategy ObjectTyper should be the provided scheme")
	}
}
