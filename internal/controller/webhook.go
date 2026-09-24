package controller

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/pathlyhq/pathly-crossplane/apis/pathly/v1alpha1"
	pathlyclient "github.com/pathlyhq/pathly-crossplane/internal/client"
)

const (
	errNotWebhook = "managed resource is not a Webhook"
	errObserveWh  = "cannot observe Webhook"
	errCreateWh   = "cannot create Webhook"
	errDeleteWh   = "cannot delete Webhook"
)

// SetupWebhook adds a controller that reconciles Webhook managed resources.
func SetupWebhook(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.WebhookGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.WebhookGroupVersionKind),
		managed.WithExternalConnecter(&webhookConnector{kube: mgr.GetClient()}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(o.PollInterval),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Webhook{}).
		Complete(r)
}

type webhookConnector struct {
	kube client.Client
}

func (c *webhookConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cli, err := NewPathlyClient(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	return &webhookExternal{client: cli}, nil
}

type webhookExternal struct {
	client *pathlyclient.Client
}

func (e *webhookExternal) Disconnect(_ context.Context) error { return nil }

func (e *webhookExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotWebhook)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	got, err := e.client.GetWebhook(ctx, ext)
	if err != nil {
		if pathlyclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errObserveWh)
	}
	cr.Status.AtProvider = v1alpha1.WebhookObservation{
		ID:             got.ID,
		Events:         append([]string(nil), got.Events...),
		Enabled:        CopyBoolPtr(got.Enabled),
		HasSecret:      CopyBoolPtr(got.HasSecret),
		URLFingerprint: CopyStringPtr(got.URLFingerprint),
		CreatedAt:      CopyStringPtr(got.CreatedAt),
	}
	// Webhooks cannot be updated in place via the public API.
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *webhookExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotWebhook)
	}
	in := pathlyclient.WebhookInput{
		URL:    cr.Spec.ForProvider.URL,
		Events: append([]string(nil), cr.Spec.ForProvider.Events...),
	}
	got, err := e.client.CreateWebhook(ctx, in, idempotencyKey(cr.UID, "create"))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateWh)
	}
	meta.SetExternalName(cr, got.ID)
	cd := managed.ConnectionDetails{}
	if got.Secret != nil {
		cd["secret"] = []byte(*got.Secret)
	}
	return managed.ExternalCreation{ConnectionDetails: cd}, nil
}

func (e *webhookExternal) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *webhookExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Webhook)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotWebhook)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalDelete{}, nil
	}
	err := e.client.DeleteWebhook(ctx, ext)
	if pathlyclient.IsNotFound(err) {
		return managed.ExternalDelete{}, nil
	}
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteWh)
}
