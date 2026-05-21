package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AcknowledgeOptions is the input for the /acknowledge subresource.
type AcknowledgeOptions struct {
	metav1.TypeMeta `json:",inline"`

	// Note allows the user to add context about the acknowledgement.
	// +optional
	Note string `json:"note,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// UnacknowledgeOptions is the input for the /unacknowledge subresource.
type UnacknowledgeOptions struct {
	metav1.TypeMeta `json:",inline"`

	// Note allows the user to add context about why they are unacknowledging.
	// +optional
	Note string `json:"note,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SnoozeOptions is the input for the /snooze subresource.
type SnoozeOptions struct {
	metav1.TypeMeta `json:",inline"`

	// Duration is a relative duration string (e.g., "1h", "4h", "1d", "7d").
	// Either Duration or Until must be specified.
	// +optional
	Duration string `json:"duration,omitempty"`

	// Until is an absolute time when the snooze expires.
	// Either Duration or Until must be specified.
	// +optional
	Until *metav1.Time `json:"until,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResolveOptions is the input for the /resolve subresource.
type ResolveOptions struct {
	metav1.TypeMeta `json:",inline"`

	// Note allows the user to add context about the resolution.
	// +optional
	Note string `json:"note,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AssignOptions is the input for the /assign subresource.
type AssignOptions struct {
	metav1.TypeMeta `json:",inline"`

	// Assignee is who to assign the insight to.
	Assignee ActorReference `json:"assignee"`

	// Note allows the user to add context about the assignment.
	// +optional
	Note string `json:"note,omitempty"`
}
