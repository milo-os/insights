package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InsightSeverity represents the severity level of an insight.
// +kubebuilder:validation:Enum=info;warning;critical
type InsightSeverity string

const (
	InsightSeverityInfo     InsightSeverity = "info"
	InsightSeverityWarning  InsightSeverity = "warning"
	InsightSeverityCritical InsightSeverity = "critical"
)

// InsightState represents the lifecycle state of an insight.
// +kubebuilder:validation:Enum=Active;Acknowledged;Snoozed;Resolved
type InsightState string

const (
	InsightStateActive       InsightState = "Active"
	InsightStateAcknowledged InsightState = "Acknowledged"
	InsightStateSnoozed      InsightState = "Snoozed"
	InsightStateResolved     InsightState = "Resolved"
)

// InsightSourceType represents the type of source that created the insight.
// +kubebuilder:validation:Enum=Manual;Policy
type InsightSourceType string

const (
	InsightSourceTypeManual InsightSourceType = "Manual"
	InsightSourceTypePolicy InsightSourceType = "Policy"
)

// TargetReference identifies the resource that this insight is about.
type TargetReference struct {
	// APIVersion of the target resource.
	APIVersion string `json:"apiVersion"`

	// Kind of the target resource.
	Kind string `json:"kind"`

	// Name of the target resource.
	Name string `json:"name"`

	// Namespace of the target resource. Empty for cluster-scoped resources.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// PolicyReference identifies the InsightPolicy that created this insight.
type PolicyReference struct {
	// Name of the InsightPolicy.
	Name string `json:"name"`

	// Namespace of the InsightPolicy.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// RuleName is the name of the rule within the policy that matched.
	RuleName string `json:"ruleName"`
}

// InsightSource describes the origin of the insight.
type InsightSource struct {
	// Type indicates whether this insight was created manually or by a policy.
	Type InsightSourceType `json:"type"`

	// Component identifies the component that created a manual insight.
	// +optional
	Component string `json:"component,omitempty"`

	// PolicyRef references the InsightPolicy that created this insight.
	// +optional
	PolicyRef *PolicyReference `json:"policyRef,omitempty"`
}

// ActorReference identifies a user or service account that performed an action.
type ActorReference struct {
	// Type is the actor type: "user" or "serviceaccount"
	Type string `json:"type"`

	// Name is the display name of the actor.
	Name string `json:"name"`

	// UID is the unique identifier of the actor.
	// +optional
	UID string `json:"uid,omitempty"`

	// Email is the email address for user actors.
	// +optional
	Email string `json:"email,omitempty"`
}

// AcknowledgementInfo tracks who acknowledged the insight and when.
type AcknowledgementInfo struct {
	// By is who acknowledged the insight.
	By ActorReference `json:"by"`

	// At is when the insight was acknowledged.
	At metav1.Time `json:"at"`

	// Note is an optional message provided during acknowledgement.
	// +optional
	Note string `json:"note,omitempty"`
}

// SnoozeInfo tracks snooze configuration.
type SnoozeInfo struct {
	// By is who snoozed the insight.
	By ActorReference `json:"by"`

	// At is when the insight was snoozed.
	At metav1.Time `json:"at"`

	// Until is when the snooze expires.
	Until metav1.Time `json:"until"`
}

// AssignmentInfo tracks assignment details.
type AssignmentInfo struct {
	// By is who assigned the insight.
	By ActorReference `json:"by"`

	// To is who the insight is assigned to.
	To ActorReference `json:"to"`

	// At is when the assignment was made.
	At metav1.Time `json:"at"`

	// Note is an optional message provided during assignment.
	// +optional
	Note string `json:"note,omitempty"`
}

// ResolutionInfo tracks resolution details.
type ResolutionInfo struct {
	// By is who resolved the insight.
	By ActorReference `json:"by"`

	// At is when the insight was resolved.
	At metav1.Time `json:"at"`

	// Note is an optional message provided during resolution.
	// +optional
	Note string `json:"note,omitempty"`
}

// InsightSpec defines the desired state of Insight.
type InsightSpec struct {
	// TargetRef identifies the resource this insight is about.
	TargetRef TargetReference `json:"targetRef"`

	// Severity indicates the severity level of this insight.
	Severity InsightSeverity `json:"severity"`

	// Message is a short summary of the insight.
	// +kubebuilder:validation:MaxLength=256
	Message string `json:"message"`

	// Description provides detailed information about the insight.
	// +optional
	// +kubebuilder:validation:MaxLength=4096
	Description string `json:"description,omitempty"`

	// Category classifies the type of insight.
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Category string `json:"category"`

	// Source describes the origin of this insight.
	Source InsightSource `json:"source"`

	// TTLSeconds specifies the time-to-live in seconds.
	// After this duration, the insight will be marked as expired.
	// A value of 0 means the insight never expires.
	// +optional
	// +kubebuilder:default=0
	TTLSeconds int64 `json:"ttlSeconds,omitempty"`
}

// InsightStatus defines the observed state of Insight.
type InsightStatus struct {
	// State is the current lifecycle state.
	// +optional
	State InsightState `json:"state,omitempty"`

	// Owner is the current responsible party for this insight.
	// +optional
	Owner *ActorReference `json:"owner,omitempty"`

	// Acknowledgement tracks acknowledgement details.
	// +optional
	Acknowledgement *AcknowledgementInfo `json:"acknowledgement,omitempty"`

	// Snooze tracks snooze configuration if the insight is snoozed.
	// +optional
	Snooze *SnoozeInfo `json:"snooze,omitempty"`

	// Assignment tracks the current assignment.
	// +optional
	Assignment *AssignmentInfo `json:"assignment,omitempty"`

	// Resolution tracks resolution details if resolved.
	// +optional
	Resolution *ResolutionInfo `json:"resolution,omitempty"`

	// TargetExists indicates whether the target resource still exists.
	// +optional
	TargetExists *bool `json:"targetExists,omitempty"`

	// ExpiresAt is when the TTL expires.
	// +optional
	ExpiresAt *metav1.Time `json:"expiresAt,omitempty"`

	// DeleteAt is when a resolved insight will be permanently deleted.
	// Set to 30 days after resolution.
	// +optional
	DeleteAt *metav1.Time `json:"deleteAt,omitempty"`

	// Muted indicates whether this insight is muted by a mute rule.
	// +optional
	Muted bool `json:"muted,omitempty"`

	// MutedBy references the InsightMuteRule that muted this insight.
	// +optional
	MutedBy *MuteRuleReference `json:"mutedBy,omitempty"`

	// Conditions represent the latest available observations.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recent generation observed.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// MuteRuleReference identifies an InsightMuteRule.
type MuteRuleReference struct {
	// Name of the InsightMuteRule.
	Name string `json:"name"`

	// Namespace of the InsightMuteRule.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Severity",type=string,JSONPath=`.spec.severity`
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.status.owner.name`
// +kubebuilder:printcolumn:name="Category",type=string,JSONPath=`.spec.category`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Insight represents an observation or finding about a Kubernetes resource.
type Insight struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InsightSpec   `json:"spec,omitempty"`
	Status InsightStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// InsightList contains a list of Insight.
type InsightList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Insight `json:"items"`
}
