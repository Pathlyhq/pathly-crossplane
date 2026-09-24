package controller

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	"github.com/pathlyhq/pathly-crossplane/apis/pathly/v1alpha1"
	pathlyclient "github.com/pathlyhq/pathly-crossplane/internal/client"
)

const (
	errGetProviderConfig = "cannot get ProviderConfig"
	errGetSecret         = "cannot get credentials Secret"
	errNoToken           = "credentials Secret has no token (expected keys: token or apiToken)"
)

// NewPathlyClient builds an API client from the managed resource's ProviderConfig.
func NewPathlyClient(ctx context.Context, kube client.Client, mg resource.Managed) (*pathlyclient.Client, error) {
	ref := mg.GetProviderConfigReference()
	if ref == nil {
		return nil, errors.New("no providerConfigRef set on managed resource")
	}

	pc := &v1alpha1.ProviderConfig{}
	if err := kube.Get(ctx, types.NamespacedName{Name: ref.Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetProviderConfig)
	}

	token, err := extractToken(ctx, kube, pc)
	if err != nil {
		return nil, err
	}

	baseURL := pathlyclient.DefaultBaseURL
	if pc.Spec.BaseURL != nil && strings.TrimSpace(*pc.Spec.BaseURL) != "" {
		baseURL = *pc.Spec.BaseURL
	}

	return pathlyclient.New(baseURL, token), nil
}

func extractToken(ctx context.Context, kube client.Client, pc *v1alpha1.ProviderConfig) (string, error) {
	src := pc.Spec.CredentialsSource
	if src == "" {
		src = xpv1.CredentialsSourceSecret
	}
	if src != xpv1.CredentialsSourceSecret {
		return "", fmt.Errorf("credentialsSource %q is not supported; use Secret", src)
	}
	if pc.Spec.CredentialsSecretRef == nil {
		return "", errors.New("credentialsSecretRef is required when credentialsSource is Secret")
	}

	ref := pc.Spec.CredentialsSecretRef
	secret := &corev1.Secret{}
	nn := types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
	if err := kube.Get(ctx, nn, secret); err != nil {
		return "", errors.Wrap(err, errGetSecret)
	}

	key := ref.Key
	if key == "" {
		key = "token"
	}
	raw, ok := secret.Data[key]
	if !ok || len(raw) == 0 {
		if key != "apiToken" {
			raw, ok = secret.Data["apiToken"]
		}
		if !ok || len(raw) == 0 {
			return "", errors.New(errNoToken)
		}
	}
	return strings.TrimSpace(string(raw)), nil
}

// Ptr helpers used by controllers when mapping CRD fields to the HTTP client.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func CopyStringPtr(in *string) *string {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func CopyInt64Ptr(in *int64) *int64 {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func CopyBoolPtr(in *bool) *bool {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}

func CopyFloat64Ptr(in *float64) *float64 {
	if in == nil {
		return nil
	}
	v := *in
	return &v
}
