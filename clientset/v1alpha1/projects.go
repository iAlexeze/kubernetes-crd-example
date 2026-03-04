package v1alpha1

import (
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type ProjectInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.ProjectList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Project, error)
	Create(*v1alpha1.Project) (*v1alpha1.Project, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Name() string
	Namespace() string
	RestClient() rest.Interface
	// ...
}

type projectClient struct {
	restClient rest.Interface
	ns         string
	name       string
}

var _ ProjectInterface = (*projectClient)(nil)

// Getters
func (c *projectClient) Name() string {
	return c.name
}

func (c *projectClient) Namespace() string {
	return c.ns
}

func (c *projectClient) RestClient() rest.Interface {
	return c.restClient
}

// Functions
func (c *projectClient) List(opts metav1.ListOptions) (*v1alpha1.ProjectList, error) {
	result := v1alpha1.ProjectList{}
	err := c.RestClient().
		Get().
		Namespace(c.Namespace()).
		Resource(c.Name()).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *projectClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Project, error) {
	result := v1alpha1.Project{}
	err := c.RestClient().
		Get().
		Namespace(c.Namespace()).
		Resource(c.Name()).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (c *projectClient) Create(project *v1alpha1.Project) (*v1alpha1.Project, error) {
	result := v1alpha1.Project{}
	err := c.RestClient().
		Post().
		Namespace(c.Namespace()).
		Resource(c.Name()).
		Body(project).
		Do().
		Into(&result)

	return &result, err
}

func (c *projectClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.RestClient().
		Get().
		Namespace(c.Namespace()).
		Resource(c.Name()).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}
