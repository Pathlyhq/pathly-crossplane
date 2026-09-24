package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SlaTargetParameters describe a Pathly SLA target.
type SlaTargetParameters struct {
	// MonitorID is the scenario this target measures.
	// +optional
	MonitorID *string `json:"monitorId,omitempty"`

	// Name of the target.
	// +optional
	Name *string `json:"name,omitempty"`

	// ObjectivePct is the uptime objective (e.g. 99.9).
	ObjectivePct float64 `json:"objectivePct"`

	// WindowDays is the rolling measurement window.
	WindowDays int64 `json:"windowDays"`

	// ExcludeMaintenance excludes runs that fall in a maintenance window.
	// +optional
	ExcludeMaintenance *bool `json:"excludeMaintenance,omitempty"`

	// WarnAtBudgetRatio raises a warning when this fraction of the error budget is consumed.
	// +optional
	WarnAtBudgetRatio *float64 `json:"warnAtBudgetRatio,omitempty"`

	// Enabled toggles the target.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// SlaTargetObservation is the read-only projection.
type SlaTargetObservation struct {
	ID                 string   `json:"id,omitempty"`
	MonitorID          *string  `json:"monitorId,omitempty"`
	Name               *string  `json:"name,omitempty"`
	ObjectivePct       *float64 `json:"objectivePct,omitempty"`
	WindowDays         *int64   `json:"windowDays,omitempty"`
	ExcludeMaintenance *bool    `json:"excludeMaintenance,omitempty"`
	WarnAtBudgetRatio  *float64 `json:"warnAtBudgetRatio,omitempty"`
	Enabled            *bool    `json:"enabled,omitempty"`
}

// SlaTargetSpec defines the desired state.
type SlaTargetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SlaTargetParameters `json:"forProvider"`
}

// SlaTargetStatus defines the observed state.
type SlaTargetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SlaTargetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pathly}
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations['crossplane\\.io/external-name']"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// SlaTarget is a managed Pathly SLA target.
type SlaTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlaTargetSpec   `json:"spec"`
	Status SlaTargetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SlaTargetList contains a list of SlaTarget.
type SlaTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SlaTarget `json:"items"`
}

func (mg *SlaTarget) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}
func (mg *SlaTarget) SetConditions(c ...xpv1.Condition) { mg.Status.SetConditions(c...) }
func (mg *SlaTarget) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}
func (mg *SlaTarget) SetDeletionPolicy(r xpv1.DeletionPolicy) { mg.Spec.DeletionPolicy = r }
func (mg *SlaTarget) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}
func (mg *SlaTarget) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}
func (mg *SlaTarget) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}
func (mg *SlaTarget) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}
func (mg *SlaTarget) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}
func (mg *SlaTarget) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

func init() {
	SchemeBuilder.Register(&SlaTarget{}, &SlaTargetList{})
}
