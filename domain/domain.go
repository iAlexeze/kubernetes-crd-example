package domain

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/utils"
)

type Component interface {

	// Start() starts the compnent
	Start(context.Context) error

	// Shutdown() shuts down the component gracefully
	Shutdown(context.Context)

	// Name() returns the name of the component
	Name() string
}

type Reconciler interface {
	// Reconcile handles the actual business logic for a resource
	Reconcile(ctx context.Context, key string) error

	// GroupVersionKind() returns the 'group/version, Kind=kind' for the reconciler.
	// Useful for reconciler registration and queuing
	GroupVersionKind() utils.GroupVersionKind
}
