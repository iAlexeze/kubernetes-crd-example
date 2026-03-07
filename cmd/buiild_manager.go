package main

import (
	"context"
	"fmt"
	"strings"

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
)

type startupCfg struct {
	controller *controller.Controller
	event      *event.Event
	kube       *kubeclient.Kubeclient
	manager    *manager.Manager
}

func buildManager(cfg *config.Config, ctx context.Context) *startupCfg {
	scheme, err := buildScheme()
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

	// Register CRD clients
	for _, crd := range buildCRDs(kube) {
		provider.Register(crd.obj, func(k *kubeclient.Kubeclient) (informer.GenericClient, error) {
			return k.NewClient(crd.listObj, crd.info)
		})
	}

	// Create informer factory
	infFactory := informer.NewFactory(
		provider,
		wq,
		scheme,
		cfg.Cluster().Namespace,
		cfg.Cluster().DefaultResync,
	)
	components = append(components, infFactory)

	// Registry
	reg := controller.NewRegistry()

	for _, crd := range buildCRDs(kube) {
		// 1. Create informer
		inf := infFactory.For(crd.obj, ctx)

		// 2. Create reconciler
		rec := crd.reconciler(inf, ev)

		// 3. Register in registry
		reg.Register(
			domain.FromGVKObj(crd.info.GroupVersionKind),
			crd.info,
			inf,
			rec,
		)
	}

	// controller
	ctrl := controller.NewController(
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

	fmt.Println("==========================")
	fmt.Println("REGISTERING MANAGER COMPONENTS...")
	for _, comp := range components {
		mgr.Register(comp)
		logger.Info().Msgf("[%s] component registered", comp.Name())
	}
	var names []string
	for _, comp := range components {
		names = append(names, comp.Name())
	}
	fmt.Printf("Available Components: %s\n", strings.Join(names, ", "))
	fmt.Println("==========================")

	return &startupCfg{
		event:      ev,
		controller: ctrl,
		kube:       kube,
		manager:    mgr,
	}
}
