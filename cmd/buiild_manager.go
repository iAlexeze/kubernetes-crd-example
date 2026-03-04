package main

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/config"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/health"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/informer"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/leader"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/manager"
)

func buildManager(cfg *config.Config) *manager.Manager {
	// create domain components
	var components []domain.Component

	// health server
	hs := health.NewHealthServer("projects", cfg.Health().Port)
	components = append(components, hs)

	// kube client
	kube := kubeclient.NewKubeclient(true, kubeclient.Options{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
	})
	components = append(components, kube)

	// informer
	informer := informer.NewInformer(nil, cfg.Cluster().Namespace, cfg.Cluster().DefaultResync)
	components = append(components, informer)

	// controller run
	controllerRun := func(ctx context.Context) {
		// Start the informer and wait for cache sync
		// informer.Start(ctx)
		logger.Info().Msgf("reached here")
		
		// informer.Controller().Run(ctx.Done())
	}

	// leader election
	leader := leader.NewLeaderElection(
		kube,
		controllerRun,
		leader.Options{
			Namespace:     cfg.Cluster().Namespace,
			LeaseDuration: cfg.Leader().LeaseDuration,
			RenewDeadline: cfg.Leader().RenewDeadline,
			RetryPeriod:   cfg.Leader().RetryPeriod,
		})
	components = append(components, leader)

	// Build and start manager
	mgr := manager.NewManager(cfg.Cluster().DefaultResync)

	// Register all manager components
	for _, comp := range components {
		mgr.Register(comp)
	}

	return mgr
}
