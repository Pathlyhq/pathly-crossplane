// Pathly Crossplane provider.
//
// Company: Pathly. Author: Simon Raynaud / keyral.
package main

import (
	"flag"
	"os"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/pathlyhq/pathly-crossplane/apis/pathly/v1alpha1"
	pathlycontroller "github.com/pathlyhq/pathly-crossplane/internal/controller"
)

func main() {
	var (
		debug          = flag.Bool("debug", false, "Enable verbose logging")
		syncInterval   = flag.Duration("sync", time.Hour, "How often all resources will be double-checked for drift")
		pollInterval   = flag.Duration("poll", time.Minute, "How often individual resources will be checked for drift")
		leaderElection = flag.Bool("leader-election", false, "Use leader election")
		maxReconcile   = flag.Int("max-reconcile-rate", 10, "Maximum reconcile rate")
	)
	flag.Parse()

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-pathly"))
	ctrl.SetLogger(zl)

	cfg, err := ctrl.GetConfig()
	if err != nil {
		log.Info("Cannot get API server rest config", "error", err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		LeaderElection:             *leaderElection,
		LeaderElectionID:           "crossplane-leader-election-provider-pathly",
		LeaderElectionResourceLock: resourcelock.LeasesResourceLock,
		Cache:                      cache.Options{SyncPeriod: syncInterval},
		Metrics:                    metricsserver.Options{BindAddress: ":8080"},
	})
	if err != nil {
		log.Info("Cannot create controller manager", "error", err)
		os.Exit(1)
	}

	if err := v1alpha1.AddToScheme(mgr.GetScheme()); err != nil {
		log.Info("Cannot add Pathly APIs to scheme", "error", err)
		os.Exit(1)
	}

	o := controller.Options{
		Logger:                  log,
		MaxConcurrentReconciles: *maxReconcile,
		PollInterval:            *pollInterval,
		GlobalRateLimiter:       ratelimiter.NewGlobal(*maxReconcile),
	}
	if err := pathlycontroller.Setup(mgr, o); err != nil {
		log.Info("Cannot setup Pathly controllers", "error", err)
		os.Exit(1)
	}
	log.Info("Starting Pathly provider")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Info("Cannot start controller manager", "error", err)
		os.Exit(1)
	}
}
