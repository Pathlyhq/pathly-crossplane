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
	errNotMW     = "managed resource is not a MaintenanceWindow"
	errObserveMW = "cannot observe MaintenanceWindow"
	errCreateMW  = "cannot create MaintenanceWindow"
	errDeleteMW  = "cannot delete MaintenanceWindow"
)

// SetupMaintenanceWindow adds a controller for MaintenanceWindow resources.
func SetupMaintenanceWindow(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.MaintenanceWindowGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.MaintenanceWindowGroupVersionKind),
		managed.WithExternalConnecter(&mwConnector{kube: mgr.GetClient()}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(o.PollInterval),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.MaintenanceWindow{}).
		Complete(r)
}

type mwConnector struct {
	kube client.Client
}

func (c *mwConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cli, err := NewPathlyClient(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	return &mwExternal{client: cli}, nil
}

type mwExternal struct {
	client *pathlyclient.Client
}

func (e *mwExternal) Disconnect(_ context.Context) error { return nil }

func (e *mwExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.MaintenanceWindow)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMW)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	got, err := e.client.GetMaintenanceWindow(ctx, ext)
	if err != nil {
		if pathlyclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errObserveMW)
	}
	cr.Status.AtProvider = v1alpha1.MaintenanceWindowObservation{
		ID:          got.ID,
		MonitorID:   CopyStringPtr(got.MonitorID),
		StartsAt:    CopyStringPtr(got.StartsAt),
		EndsAt:      CopyStringPtr(got.EndsAt),
		Reason:      CopyStringPtr(got.Reason),
		Weekday:     CopyInt64Ptr(got.Weekday),
		StartMinute: CopyInt64Ptr(got.StartMinute),
		DurationMin: CopyInt64Ptr(got.DurationMin),
	}
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *mwExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.MaintenanceWindow)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMW)
	}
	p := cr.Spec.ForProvider
	in := pathlyclient.MaintenanceWindowInput{
		MonitorID:   CopyStringPtr(p.MonitorID),
		StartsAt:    CopyStringPtr(p.StartsAt),
		EndsAt:      CopyStringPtr(p.EndsAt),
		Reason:      CopyStringPtr(p.Reason),
		Weekday:     CopyInt64Ptr(p.Weekday),
		StartMinute: CopyInt64Ptr(p.StartMinute),
		DurationMin: CopyInt64Ptr(p.DurationMin),
	}
	got, err := e.client.CreateMaintenanceWindow(ctx, in, idempotencyKey(cr.UID, "create"))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateMW)
	}
	meta.SetExternalName(cr, got.ID)
	return managed.ExternalCreation{}, nil
}

func (e *mwExternal) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

func (e *mwExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.MaintenanceWindow)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotMW)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalDelete{}, nil
	}
	err := e.client.DeleteMaintenanceWindow(ctx, ext)
	if pathlyclient.IsNotFound(err) {
		return managed.ExternalDelete{}, nil
	}
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteMW)
}
