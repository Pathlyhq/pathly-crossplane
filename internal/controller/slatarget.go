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
	errNotSla     = "managed resource is not a SlaTarget"
	errObserveSla = "cannot observe SlaTarget"
	errCreateSla  = "cannot create SlaTarget"
	errUpdateSla  = "cannot update SlaTarget"
	errDeleteSla  = "cannot delete SlaTarget"
)

// SetupSlaTarget adds a controller for SlaTarget resources.
func SetupSlaTarget(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.SlaTargetGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.SlaTargetGroupVersionKind),
		managed.WithExternalConnecter(&slaConnector{kube: mgr.GetClient()}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(o.PollInterval),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.SlaTarget{}).
		Complete(r)
}

type slaConnector struct {
	kube client.Client
}

func (c *slaConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cli, err := NewPathlyClient(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	return &slaExternal{client: cli}, nil
}

type slaExternal struct {
	client *pathlyclient.Client
}

func (e *slaExternal) Disconnect(_ context.Context) error { return nil }

func (e *slaExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.SlaTarget)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSla)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	got, err := e.client.GetSlaTarget(ctx, ext)
	if err != nil {
		if pathlyclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errObserveSla)
	}
	cr.Status.AtProvider = v1alpha1.SlaTargetObservation{
		ID:                 got.ID,
		MonitorID:          CopyStringPtr(got.MonitorID),
		Name:               CopyStringPtr(got.Name),
		ObjectivePct:       CopyFloat64Ptr(got.ObjectivePct),
		WindowDays:         CopyInt64Ptr(got.WindowDays),
		ExcludeMaintenance: CopyBoolPtr(got.ExcludeMaintenance),
		WarnAtBudgetRatio:  CopyFloat64Ptr(got.WarnAtBudgetRatio),
		Enabled:            CopyBoolPtr(got.Enabled),
	}
	upToDate := true
	if got.ObjectivePct == nil || *got.ObjectivePct != cr.Spec.ForProvider.ObjectivePct {
		upToDate = false
	}
	if got.WindowDays == nil || *got.WindowDays != cr.Spec.ForProvider.WindowDays {
		upToDate = false
	}
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *slaExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SlaTarget)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSla)
	}
	got, err := e.client.UpsertSlaTarget(ctx, slaInput(cr.Spec.ForProvider), idempotencyKey(cr.UID, "create"))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateSla)
	}
	meta.SetExternalName(cr, got.ID)
	return managed.ExternalCreation{}, nil
}

func (e *slaExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.SlaTarget)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSla)
	}
	_, err := e.client.UpsertSlaTarget(ctx, slaInput(cr.Spec.ForProvider), idempotencyKey(cr.UID, "update"))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSla)
}

func (e *slaExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.SlaTarget)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSla)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalDelete{}, nil
	}
	err := e.client.DeleteSlaTarget(ctx, ext)
	if pathlyclient.IsNotFound(err) {
		return managed.ExternalDelete{}, nil
	}
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteSla)
}

func slaInput(p v1alpha1.SlaTargetParameters) pathlyclient.SlaTargetInput {
	return pathlyclient.SlaTargetInput{
		MonitorID:          CopyStringPtr(p.MonitorID),
		Name:               CopyStringPtr(p.Name),
		ObjectivePct:       p.ObjectivePct,
		WindowDays:         p.WindowDays,
		ExcludeMaintenance: CopyBoolPtr(p.ExcludeMaintenance),
		WarnAtBudgetRatio:  CopyFloat64Ptr(p.WarnAtBudgetRatio),
		Enabled:            CopyBoolPtr(p.Enabled),
	}
}
