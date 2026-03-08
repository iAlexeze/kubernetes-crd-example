package main

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/config"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/controller"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/event"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/health"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/informer"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/manager"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/registry"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/utils"
)

type startupCfg struct {
	controller *controller.Controller
	event      *event.Event
	kube       *kubeclient.Kubeclient
	manager    *manager.Manager
}

func buildManager(cfg *config.Config, ctx context.Context) *startupCfg {
	// crd registry
	crdRegistry := registry.NewCRDRegistry()

	// scheme registry
	scheme, err := registry.NewSchemeRegistry()
	if err != nil {
		logger.Fatal().Err(err).Msg("scheme creation error")
	}

	// Initialize components
	var components []domain.Component

	// health
	hs := health.NewHealthServer(cfg)
	components = append(components, hs)

	// kube
	kube := kubeclient.NewKubeclient(kubeclient.Config{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
		Scheme:     scheme,
	})
	components = append(components, kube)

	// events
	ev := event.NewEvent(kube)
	components = append(components, ev)

	// queue
	wq := queue.NewWorkqueue()
	components = append(components, wq)

	// provider
	provider := kube.ClientProvider()

	// Register CRD clients to provider - for automatic client  and informer generation
	for _, crd := range crdRegistry {
		provider.Register(crd.Object, func(k *kubeclient.Kubeclient) (informer.GenericClient, error) {
			return k.NewClient(crd.ListObject, kubeclient.CRDInfo(crd.Info))
		})
	}

	// Create shared informer factory
	infFactory := informer.SharedInformerFactory(
		provider,
		wq,
		scheme,
		cfg.Cluster().Namespace,
		cfg.Cluster().DefaultResync,
	)
	components = append(components, infFactory)

	// Controller Registry
	reg := controller.NewControllerRegistry()

	// Register CRDs to controller registry
	logger.Info().Msg("registering CRDs...")
	for _, crd := range crdRegistry {
		// 1. Create informer
		inf := infFactory.For(crd.Object, ctx)

		// 2. Create reconciler
		rec := crd.Reconciler(kube, inf, ev)

		// 3. Register in controller registry
		logger.Debug().Str("GVK", utils.SetGroupVersionKindObj(crd.Info.GroupVersionKind)).Msg("registering CRD")
		reg.Register(
			utils.SetGroupVersionKindObj(crd.Info.GroupVersionKind),
			crd.Info,
			inf,
			rec,
		)
	}

	// controller manager
	ctrl := controller.NewControllerManager(
		kube,
		infFactory,
		reg,
		ev,
		wq,
		cfg.Cluster().Workers,
	)
	components = append(components, ctrl)

	// manager
	mgr := manager.NewManager(hs, cfg.Cluster().DefaultResync)
	mgr.Register(components) // Register all manager components

	return &startupCfg{
		event:      ev,
		controller: ctrl,
		kube:       kube,
		manager:    mgr,
	}
}
