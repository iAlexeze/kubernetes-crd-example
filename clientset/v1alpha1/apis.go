package v1alpha1

import (
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
)

// Projects implements the project interface
func (p *projectClient) Projects(namespace string) domain.ProjectInterface {
	return &projectClient{
		restClient: p.restClient,
		namespace:  namespace,
		name:       "projects",
	}
}

// API Functions
func (p *projectClient) List(opts metav1.ListOptions) (*v1alpha1.ProjectList, error) {
	result := v1alpha1.ProjectList{}
	err := p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (p *projectClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Project, error) {
	result := v1alpha1.Project{}
	err := p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)

	return &result, err
}

func (p *projectClient) Create(project *v1alpha1.Project) (*v1alpha1.Project, error) {
	result := v1alpha1.Project{}
	err := p.restClient.
		Post().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		Body(project).
		Do().
		Into(&result)

	return &result, err
}

func (p *projectClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return p.restClient.
		Get().
		Namespace(p.Namespace()).
		Resource(p.Name()).
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}
