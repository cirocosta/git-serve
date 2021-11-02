package main

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware-labs/reconciler-runtime/reconcilers"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/cirocosta/git-serve/pkg"
	"github.com/cirocosta/git-serve/pkg/apis/v1alpha1"
	"github.com/cirocosta/git-serve/pkg/controllers"
)

var (
	scheme     = k8sruntime.NewScheme()
	syncPeriod = 1 * time.Hour
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func run(ctx context.Context) error {
	restConfig, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	mgr, err := manager.New(restConfig, manager.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: ":8081",
		SyncPeriod:             &syncPeriod,
	})
	if err != nil {
		return fmt.Errorf("manager new: %w", err)
	}

	err = controllers.GitServerReconciler(reconcilers.NewConfig(
		mgr, &v1alpha1.GitServer{}, syncPeriod,
	)).SetupWithManager(ctx, mgr)
	if err != nil {
		return fmt.Errorf("gitserver reconciler setupwithmgr: %w", err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("add healthz: %w", err)
	}

	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("add readyz: %w", err)
	}

	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("mgr start: %w", err)
	}

	return nil
}

func main() {
	ctx, cancel := context.WithCancel(
		pkg.SignalHandlingContext(context.Background()),
	)
	defer cancel()

	if err := run(ctx); err != nil {
		panic(err)
	}
}
