// Package v1alpha1 contains the Pathly Crossplane API types.
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	// Group is the API group for Pathly managed resources.
	Group = "pathly.pathlyhq.com"
	// Version is the API version.
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// GroupVersionKinds for managed resources.
var (
	ProviderConfigKind             = "ProviderConfig"
	ProviderConfigGroupKind        = schema.GroupKind{Group: Group, Kind: ProviderConfigKind}.String()
	ProviderConfigGroupVersionKind = SchemeGroupVersion.WithKind(ProviderConfigKind)

	ScenarioKind             = "Scenario"
	ScenarioGroupKind        = schema.GroupKind{Group: Group, Kind: ScenarioKind}.String()
	ScenarioGroupVersionKind = SchemeGroupVersion.WithKind(ScenarioKind)

	WebhookKind             = "Webhook"
	WebhookGroupKind        = schema.GroupKind{Group: Group, Kind: WebhookKind}.String()
	WebhookGroupVersionKind = SchemeGroupVersion.WithKind(WebhookKind)

	MaintenanceWindowKind             = "MaintenanceWindow"
	MaintenanceWindowGroupKind        = schema.GroupKind{Group: Group, Kind: MaintenanceWindowKind}.String()
	MaintenanceWindowGroupVersionKind = SchemeGroupVersion.WithKind(MaintenanceWindowKind)

	SlaTargetKind             = "SlaTarget"
	SlaTargetGroupKind        = schema.GroupKind{Group: Group, Kind: SlaTargetKind}.String()
	SlaTargetGroupVersionKind = SchemeGroupVersion.WithKind(SlaTargetKind)
)
