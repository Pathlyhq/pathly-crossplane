package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// MaintenanceWindowParameters describe a Pathly maintenance window.
type MaintenanceWindowParameters struct {
	// MonitorID is the scenario this window applies to. Omit for org-wide.
	// +optional
	MonitorID *string `json:"monitorId,omitempty"`

	// StartsAt is an absolute start (RFC3339). Mutually exclusive with weekday recurrence.
	// +optional
	StartsAt *string `json:"startsAt,omitempty"`

	// EndsAt is an absolute end (RFC3339).
	// +optional
	EndsAt *string `json:"endsAt,omitempty"`

	// Reason is shown in the console and alerts.
	// +optional
	Reason *string `json:"reason,omitempty"`

	// Weekday is 0=Sunday … 6=Saturday for weekly recurrence.
	// +optional
	Weekday *int64 `json:"weekday,omitempty"`

	// StartMinute is minutes from midnight for weekly recurrence.
	// +optional
	StartMinute *int64 `json:"startMinute,omitempty"`

	// DurationMin is the length of a weekly window in minutes.
	// +optional
	DurationMin *int64 `json:"durationMin,omitempty"`
}

// MaintenanceWindowObservation is the read-only projection.
type MaintenanceWindowObservation struct {
	ID          string  `json:"id,omitempty"`
	MonitorID   *string `json:"monitorId,omitempty"`
	StartsAt    *string `json:"startsAt,omitempty"`
	EndsAt      *string `json:"endsAt,omitempty"`
	Reason      *string `json:"reason,omitempty"`
	Weekday     *int64  `json:"weekday,omitempty"`
	StartMinute *int64  `json:"startMinute,omitempty"`
	DurationMin *int64  `json:"durationMin,omitempty"`
}

// MaintenanceWindowSpec defines the desired state.
type MaintenanceWindowSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       MaintenanceWindowParameters `json:"forProvider"`
}

// MaintenanceWindowStatus defines the observed state.
type MaintenanceWindowStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          MaintenanceWindowObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pathly}
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations['crossplane\\.io/external-name']"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// MaintenanceWindow is a managed Pathly maintenance window.
type MaintenanceWindow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MaintenanceWindowSpec   `json:"spec"`
	Status MaintenanceWindowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MaintenanceWindowList contains a list of MaintenanceWindow.
type MaintenanceWindowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MaintenanceWindow `json:"items"`
}

func (mg *MaintenanceWindow) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}
func (mg *MaintenanceWindow) SetConditions(c ...xpv1.Condition) { mg.Status.SetConditions(c...) }
func (mg *MaintenanceWindow) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}
func (mg *MaintenanceWindow) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}
func (mg *MaintenanceWindow) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}
func (mg *MaintenanceWindow) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}
func (mg *MaintenanceWindow) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}
func (mg *MaintenanceWindow) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}
func (mg *MaintenanceWindow) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}
func (mg *MaintenanceWindow) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

func init() {
	SchemeBuilder.Register(&MaintenanceWindow{}, &MaintenanceWindowList{})
}
