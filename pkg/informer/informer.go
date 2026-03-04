package informer

import (
	"context"
	"time"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type informer struct {
	name       string
	namespace  string
	clientSet  domain.ProjectsV1Alpha1nterface
	resync     time.Duration
	store      cache.Store
	controller cache.Controller
	queue      cache.Queue
}

var _ domain.Component = (*informer)(nil)

func NewInformer(
	clientSet domain.ProjectsV1Alpha1nterface,
	namespace string,
	resync time.Duration,
) *informer {
	return &informer{
		name:      "smart informer",
		namespace: namespace,
		clientSet: clientSet,
		resync:    resync,
	}
}

func (i *informer) Start(ctx context.Context) error {
	i.store, i.controller = i.watchResources()

	return nil
}

func (i *informer) watchResources() (cache.Store, cache.Controller) {
	return cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(lo metav1.ListOptions) (result runtime.Object, err error) {
				return i.clientSet.Projects(i.namespace).List(lo)
			},
			WatchFunc: func(lo metav1.ListOptions) (watch.Interface, error) {
				return i.clientSet.Projects(i.namespace).Watch(lo)
			},
		},
		&v1alpha1.Project{},
		i.resync,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { i.enqueue(obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { i.enqueue(newObj) },
			DeleteFunc: func(obj interface{}) { i.enqueue(obj) },
		},
	)
}

func (i *informer) enqueue(obj interface{}) {}

func (i *informer) Shutdown(ctx context.Context) {}

// Methods
func (i *informer) Controller() cache.Controller {
	return i.controller
}

func (i *informer) Store() cache.Store {
	return i.store
}

func (i *informer) Name() string {
	return i.name
}
