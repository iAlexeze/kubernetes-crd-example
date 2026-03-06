package informer

import (
	"context"
	"time"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"k8s.io/client-go/tools/cache"
)

type InformerComponents interface {
	Store() cache.Store
}

// NewProjectInformer() returns a new ProjectInformer
func NewProjectInformer(
	client domain.ProjectsV1Alpha1nterface,
	queue *queue.Queue,
	namespace  string,
	resync time.Duration,
) *ProjectInformer {
	return &ProjectInformer{
		client: client,
		Informer: Informer{
			name: string(domain.ProjectResource),
			namespace: namespace,
			queue: *queue,
			resync: resync,
		},
	}
}

// NewManagedNamespaceInformer() returns a new anagedNamespaceInformer informer
func NewManagedNamespaceInformer(
	client domain.ManagedNamespaceV1Alpha1nterface,
	queue *queue.Queue,
	namespace  string,
	resync time.Duration,
) *ManagedNamespaceInformer {
	return &ManagedNamespaceInformer{
		client: client,
		Informer: Informer{
			name: string(domain.ProjectResource),
			namespace: namespace,
			queue: *queue,
			resync: resync,
		},
	}
}


func (i *Informer) Start(ctx context.Context) error {
	return nil
}

// enqueue adds the object's key to the workqueue
// Shutdown gracefully stops the informer
func (i *Informer) Shutdown(ctx context.Context) {}

