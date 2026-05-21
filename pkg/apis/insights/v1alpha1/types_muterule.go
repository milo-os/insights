package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InsightMuteMatch defines criteria for matching insights to mute.
type InsightMuteMatch struct {
	// PolicyRef mutes insights from a specific policy.
	// +optional
	PolicyRef *PolicyReference `json:"policyRef,omitempty"`

	// Category mutes insights of a specific category.
	// +optional
	Category string `json:"category,omitempty"`

	// TargetRef mutes insights about a specific resource.
	// +optional
	TargetRef *TargetReference `json:"targetRef,omitempty"`

	// Severity mutes insights at or below a specific severity level.
	// For example, "warning" would mute "info" and "warning" but not "critical".
	// +optional
	Severity *InsightSeverity `json:"severity,omitempty"`

	// LabelSelector mutes insights matching the specified labels.
	// +optional
	LabelSelector *metav1.LabelSelector `json:"labelSelector,omitempty"`
}

// InsightMuteRuleSpec defines the desired state of InsightMuteRule.
type InsightMuteRuleSpec struct {
	// Match defines which insights this rule mutes.
	Match InsightMuteMatch `json:"match"`

	// Reason explains why these insights are muted.
	// +optional
	// +kubebuilder:validation:MaxLength=256
	Reason string `json:"reason,omitempty"`

	// ExpiresAt allows temporary mutes that auto-expire.
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`
}

// InsightMuteRuleStatus defines the observed state of InsightMuteRule.
type InsightMuteRuleStatus struct {
	// MutedInsightCount is the number of insights currently muted by this rule.
	// +optional
	MutedInsightCount int32 `json:"mutedInsightCount,omitempty"`

	// Expired indicates whether this mute rule has expired.
	// +optional
	Expired bool `json:"expired,omitempty"`

	// Conditions represent the latest available observations.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recent generation observed.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Category",type=string,JSONPath=`.spec.match.category`
// +kubebuilder:printcolumn:name="Severity",type=string,JSONPath=`.spec.match.severity`
// +kubebuilder:printcolumn:name="Muted",type=integer,JSONPath=`.status.mutedInsightCount`
// +kubebuilder:printcolumn:name="Expires",type=date,JSONPath=`.spec.expiresAt`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// InsightMuteRule allows consumers to suppress insights matching certain criteria.
type InsightMuteRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InsightMuteRuleSpec   `json:"spec,omitempty"`
	Status InsightMuteRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// InsightMuteRuleList contains a list of InsightMuteRule.
type InsightMuteRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InsightMuteRule `json:"items"`
}
