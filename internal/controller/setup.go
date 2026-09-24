package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

// Setup registers all Pathly controllers with the manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return setupAll(mgr, o,
		SetupScenario,
		SetupWebhook,
		SetupMaintenanceWindow,
		SetupSlaTarget,
	)
}

func setupAll(mgr ctrl.Manager, o controller.Options, fns ...func(ctrl.Manager, controller.Options) error) error {
	for _, setup := range fns {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}

// DefaultOptions returns controller options with the provided logger.
func DefaultOptions(log logging.Logger) controller.Options {
	return controller.Options{Logger: log}
}
