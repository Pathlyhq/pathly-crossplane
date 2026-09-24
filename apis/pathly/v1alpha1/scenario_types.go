package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ScenarioParameters describe the desired Pathly HTTP scenario.
//
// HTTP-only policy: only type "http" is accepted. Browser journeys carry
// credentials that must not land in etcd; manage those at https://pathlyhq.com
// (console or API), never through Crossplane.
type ScenarioParameters struct {
	// Name is the human-readable scenario name.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Type of the scenario. Only "http" is accepted (HTTP-only policy —
	// browser step ASTs are rejected so credentials never land in etcd).
	// See https://pathlyhq.com/en/developers
	// +kubebuilder:validation:Enum=http
	// +kubebuilder:default=http
	Type string `json:"type"`

	// URL to probe.
	// +optional
	URL *string `json:"url,omitempty"`

	// IntervalSec between runs.
	// +optional
	IntervalSec *int64 `json:"intervalSec,omitempty"`

	// Method is the HTTP method (GET, POST, …).
	// +optional
	Method *string `json:"method,omitempty"`

	// ExpectedStatus is the HTTP status code that means success.
	// +optional
	ExpectedStatus *int64 `json:"expectedStatus,omitempty"`

	// MaxLatencyMs fails the run when the response is slower.
	// +optional
	MaxLatencyMs *int64 `json:"maxLatencyMs,omitempty"`

	// ExpectText fails the run when the body does not contain this string.
	// +optional
	ExpectText *string `json:"expectText,omitempty"`

	// Regions where the probe runs.
	// +optional
	Regions []string `json:"regions,omitempty"`

	// Tags for organization.
	// +optional
	Tags []string `json:"tags,omitempty"`

	// Folder groups scenarios in the console.
	// +optional
	Folder *string `json:"folder,omitempty"`

	// Severity of alerts raised by this scenario.
	// +optional
	Severity *string `json:"severity,omitempty"`

	// Runbook URL shown in alerts.
	// +optional
	Runbook *string `json:"runbook,omitempty"`

	// Cron expression for schedule-based runs.
	// +optional
	Cron *string `json:"cron,omitempty"`

	// Enabled toggles the scenario.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// ScenarioObservation is the read-only projection returned by the API.
type ScenarioObservation struct {
	// ID is the Pathly scenario identifier.
	ID string `json:"id,omitempty"`
	// LastStatus is the outcome of the latest run.
	LastStatus *string `json:"lastStatus,omitempty"`
	// ScenarioFingerprint is a hash of the scenario definition.
	ScenarioFingerprint *string `json:"scenarioFingerprint,omitempty"`
	// CreatedAt is the creation timestamp from Pathly.
	CreatedAt *string `json:"createdAt,omitempty"`
	// MutedUntil is set when the scenario is muted.
	MutedUntil *string `json:"mutedUntil,omitempty"`
}

// ScenarioSpec defines the desired state of Scenario.
type ScenarioSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ScenarioParameters `json:"forProvider"`
}

// ScenarioStatus defines the observed state of Scenario.
type ScenarioStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ScenarioObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,pathly}
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations['crossplane\\.io/external-name']"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// Scenario is a managed Pathly HTTP scenario (HTTP-only; no browser journeys).
// Create tokens at https://pathlyhq.com — store PATHLY_API_TOKEN in a Secret
// referenced by ProviderConfig.
type Scenario struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScenarioSpec   `json:"spec"`
	Status ScenarioStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScenarioList contains a list of Scenario.
type ScenarioList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Scenario `json:"items"`
}

// GetCondition of this Scenario.
func (mg *Scenario) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions of this Scenario.
func (mg *Scenario) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetDeletionPolicy of this Scenario.
func (mg *Scenario) GetDeletionPolicy() xpv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// SetDeletionPolicy of this Scenario.
func (mg *Scenario) SetDeletionPolicy(r xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// GetManagementPolicies of this Scenario.
func (mg *Scenario) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies of this Scenario.
func (mg *Scenario) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// GetProviderConfigReference of this Scenario.
func (mg *Scenario) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// SetProviderConfigReference of this Scenario.
func (mg *Scenario) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// GetWriteConnectionSecretToReference of this Scenario.
func (mg *Scenario) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetWriteConnectionSecretToReference of this Scenario.
func (mg *Scenario) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

func init() {
	SchemeBuilder.Register(&Scenario{}, &ScenarioList{})
}
