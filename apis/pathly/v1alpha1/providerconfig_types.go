package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ProviderConfigSpec defines how the Pathly provider authenticates.
type ProviderConfigSpec struct {
	// CredentialsSource selects how credentials are obtained.
	// +kubebuilder:validation:Enum=Secret;None;InjectedIdentity;Environment;Filesystem
	// +kubebuilder:default=Secret
	CredentialsSource xpv1.CredentialsSource `json:"credentialsSource"`

	// CredentialsSecretRef points at a Secret that holds the Pathly API token
	// under the key `token` (or `apiToken`). Required when credentialsSource is Secret.
	// +optional
	CredentialsSecretRef *xpv1.SecretKeySelector `json:"credentialsSecretRef,omitempty"`

	// BaseURL overrides the Pathly API base URL. Defaults to https://api.pathlyhq.com.
	// +optional
	BaseURL *string `json:"baseURL,omitempty"`
}

// ProviderConfigStatus reports readiness of the configuration.
type ProviderConfigStatus struct {
	xpv1.ProviderConfigStatus `json:",inline"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,pathly}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentialsSecretRef.name",priority=1

// ProviderConfig configures credentials for Pathly managed resources.
type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderConfigSpec   `json:"spec"`
	Status ProviderConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProviderConfigList contains a list of ProviderConfig.
type ProviderConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProviderConfig `json:"items"`
}

// GetUsers returns the ProviderConfig usage count.
func (p *ProviderConfig) GetUsers() int64 {
	return p.Status.Users
}

// SetUsers sets the ProviderConfig usage count.
func (p *ProviderConfig) SetUsers(i int64) {
	p.Status.Users = i
}

func init() {
	SchemeBuilder.Register(&ProviderConfig{}, &ProviderConfigList{})
}
