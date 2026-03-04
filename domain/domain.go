package domain

import "context"

type Component interface {

	// Start() starts the compnent
	Start(context.Context) error

	// Shutdown() shuts down the component gracefully
	Shutdown(context.Context)

	// Name() returns the name of the component
	Name() string
}

// To implement, replace 'componentName' with the appropriate component
// var _ domain.Component = (*componentName)(nil)
// func (c *componentName) Start(ctx context.Context) error {}
// func (c *componentName) Shutdown(ctx context.Context) {}
// func (c *componentName) Name() string {}
