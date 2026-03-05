package main

import (
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/config"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/controller"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/events"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/health"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/informer"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	clientV1alpha1 "github.com/ialexeze/kubernetes-crd-example/pkg/config/clientset/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/manager"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

type startup struct {
	events     *events.Recorder
	controller *controller.Controller
	kube       *kubeclient.Kubeclient
	manager    *manager.Manager
}

func buildManager(cfg *config.Config) *startup {
	// ── Add scheme ─────────────────────────────────────────────────────────────
	// Register both built-in types and the CRD types.
	// The scheme is needed by the CRD informer to decode API responses
	// into typed Go structs (Example *Project).
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		logger.Fatal().Err(err).Msg("failed to add client-go scheme")
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		logger.Fatal().Err(err).Msg("failed to add CRD scheme")
	}

	// create domain components
	var components []domain.Component

	// health server
	hs := health.NewHealthServer("projects", cfg)
	components = append(components, hs)

	// kube client
	kube := kubeclient.NewKubeclient(true, kubeclient.Options{
		Kubeconfig: cfg.Cluster().KubeconfigPath,
		Masterurl:  cfg.Cluster().MasterURL,
		Scheme:     scheme,
	})
	components = append(components, kube)

	// projects
	projects := clientV1alpha1.NewProjectClient(kube, scheme, cfg.Cluster().Namespace)
	components = append(components, projects)

	// informer
	informer := informer.NewInformer(projects, cfg.Cluster().DefaultResync)
	components = append(components, informer)

	// events
	events := events.NewRecorder(kube, scheme, events.Options{Component: cfg.App().Name})
	components = append(components, events)

	// controller
	ctrl := controller.NewController(
		kube,
		informer,
		events,
		cfg.Cluster().Workers,
		controller.CustomOptions{
			IsCustom: true,
			Group:    v1alpha1.GroupName,
			Kind:     v1alpha1.GroupKind,
			Version:  v1alpha1.GroupVersion,
		},
	)
	components = append(components, ctrl) // Needed to get the controller informer synced and ready for manager to finish infrastructure setup

	// Build and start manager
	mgr := manager.NewManager(hs, cfg.Cluster().DefaultResync)

	// Register all manager components
	for _, comp := range components {
		mgr.Register(comp)
	}

	return &startup{
		events:     events,
		controller: ctrl,
		kube:       kube,
		manager:    mgr,
	}
}
