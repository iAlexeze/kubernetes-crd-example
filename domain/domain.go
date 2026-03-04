package domain

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

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

type ProjectInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.ProjectList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Project, error)
	Create(*v1alpha1.Project) (*v1alpha1.Project, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Namespace() string
	// ...
}

type ProjectsV1Alpha1nterface interface {
	Projects(namespace string) ProjectInterface
}
