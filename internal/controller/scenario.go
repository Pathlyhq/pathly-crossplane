package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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
	errNotScenario = "managed resource is not a Scenario"
	errObserveScen = "cannot observe Scenario"
	errCreateScen  = "cannot create Scenario"
	errUpdateScen  = "cannot update Scenario"
	errDeleteScen  = "cannot delete Scenario"
)

// SetupScenario adds a controller that reconciles Scenario managed resources.
func SetupScenario(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.ScenarioGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.ScenarioGroupVersionKind),
		managed.WithExternalConnecter(&scenarioConnector{kube: mgr.GetClient()}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(o.PollInterval),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Scenario{}).
		Complete(r)
}

type scenarioConnector struct {
	kube client.Client
}

func (c *scenarioConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cli, err := NewPathlyClient(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	return &scenarioExternal{client: cli}, nil
}

type scenarioExternal struct {
	client *pathlyclient.Client
}

func (e *scenarioExternal) Disconnect(_ context.Context) error { return nil }

func (e *scenarioExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Scenario)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotScenario)
	}

	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	got, err := e.client.GetScenario(ctx, ext)
	if err != nil {
		if pathlyclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errObserveScen)
	}

	cr.Status.AtProvider = v1alpha1.ScenarioObservation{
		ID:                  got.ID,
		LastStatus:          CopyStringPtr(got.LastStatus),
		ScenarioFingerprint: CopyStringPtr(got.ScenarioFingerprint),
		CreatedAt:           CopyStringPtr(got.CreatedAt),
		MutedUntil:          CopyStringPtr(got.MutedUntil),
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  scenarioUpToDate(cr.Spec.ForProvider, got),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *scenarioExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Scenario)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotScenario)
	}

	in := scenarioInput(cr.Spec.ForProvider)
	idem := idempotencyKey(cr.UID, "create")
	got, err := e.client.CreateScenario(ctx, in, idem)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateScen)
	}
	meta.SetExternalName(cr, got.ID)
	return managed.ExternalCreation{}, nil
}

func (e *scenarioExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Scenario)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotScenario)
	}
	ext := meta.GetExternalName(cr)
	in := scenarioInput(cr.Spec.ForProvider)
	_, err := e.client.UpdateScenario(ctx, ext, in)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateScen)
}

func (e *scenarioExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Scenario)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotScenario)
	}
	ext := meta.GetExternalName(cr)
	if ext == "" {
		return managed.ExternalDelete{}, nil
	}
	err := e.client.DeleteScenario(ctx, ext)
	if pathlyclient.IsNotFound(err) {
		return managed.ExternalDelete{}, nil
	}
	return managed.ExternalDelete{}, errors.Wrap(err, errDeleteScen)
}

func scenarioInput(p v1alpha1.ScenarioParameters) pathlyclient.ScenarioInput {
	typ := p.Type
	if typ == "" {
		typ = "http"
	}
	name := p.Name
	return pathlyclient.ScenarioInput{
		Name:           &name,
		Type:           &typ,
		URL:            CopyStringPtr(p.URL),
		IntervalSec:    CopyInt64Ptr(p.IntervalSec),
		Method:         CopyStringPtr(p.Method),
		ExpectedStatus: CopyInt64Ptr(p.ExpectedStatus),
		MaxLatencyMs:   CopyInt64Ptr(p.MaxLatencyMs),
		ExpectText:     CopyStringPtr(p.ExpectText),
		Regions:        append([]string(nil), p.Regions...),
		Tags:           append([]string(nil), p.Tags...),
		Folder:         CopyStringPtr(p.Folder),
		Severity:       CopyStringPtr(p.Severity),
		Runbook:        CopyStringPtr(p.Runbook),
		Cron:           CopyStringPtr(p.Cron),
		Enabled:        CopyBoolPtr(p.Enabled),
	}
}

func scenarioUpToDate(want v1alpha1.ScenarioParameters, got *pathlyclient.Scenario) bool {
	if got.Name != want.Name {
		return false
	}
	if want.URL != nil && (got.URL == nil || *got.URL != *want.URL) {
		return false
	}
	if want.Enabled != nil && (got.Enabled == nil || *got.Enabled != *want.Enabled) {
		return false
	}
	if want.IntervalSec != nil && (got.IntervalSec == nil || *got.IntervalSec != *want.IntervalSec) {
		return false
	}
	if want.Method != nil && (got.Method == nil || *got.Method != *want.Method) {
		return false
	}
	if want.ExpectedStatus != nil && (got.ExpectedStatus == nil || *got.ExpectedStatus != *want.ExpectedStatus) {
		return false
	}
	return true
}

func idempotencyKey(uid interface{}, action string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%v:%s", uid, action)))
	return hex.EncodeToString(h[:16])
}
