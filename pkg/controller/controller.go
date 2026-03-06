package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/event"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/informer"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/queue"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kube        *kubeclient.Kubeclient
	informers   []informer.InformerComponents
	event       *event.Event
	queue       workqueue.TypedRateLimitingInterface[queue.QueueItem]
	wg          sync.WaitGroup
	workers     int
	reconcilers []domain.Reconciler
	opts        CustomOptions
}

var _ domain.Component = (*Controller)(nil)

func NewController(
	kube *kubeclient.Kubeclient,
	informers []informer.InformerComponents,
	event *event.Event,
	queue *queue.Queue,
	workers int,
	opts CustomOptions,
) *Controller {
	return &Controller{
		kube:      kube,
		informers: informers,
		event:     event,
		workers:   workers,
		opts:      opts,
	}
}

type CustomOptions struct {
	IsCustom bool
	Group    string
	Kind     string
	Version  string
}

func (c *Controller) Start(ctx context.Context) error {
	// Confirm CRD type presence in cluster if custom
	if c.opts.IsCustom {
		logger.Info().Msg("Custom controller setup detected")
		required := map[string]string{
			"Group":   c.opts.Group,
			"Kind":    c.opts.Kind,
			"Version": c.opts.Version,
		}

		if err := utils.RequireStrParams(required); err != nil {
			return err
		}

		// Try with backoff
		logger.Info().
			Msgf("checking %s CRD: %s/%s...", c.opts.Kind, c.opts.Group, c.opts.Version)
		if err := utils.RetryBackoff(
			func() error {
				return utils.WaitForCRD(
					c.kube.RestConfig(),
					c.opts.Group,
					c.opts.Kind,
					c.opts.Version,
				)
			}, 5, 2*time.Second,
		); err != nil {
			logger.Error().Err(err).
				Msgf("%s CRD: %s/%s... not found", c.opts.Kind, c.opts.Group, c.opts.Version)
			return err
		}

		logger.Info().
			Msgf("Found %s CRD: %s/%s...", c.opts.Kind, c.opts.Group, c.opts.Version)
	}

	if informer == nil {
		return fmt.Errorf("controller error: informer not initialized")
	}

	ctrl := informer.Controller()

	// Start the Controller
	logger.Debug().Msg("starting controller...")
	go ctrl.Run(wait.NeverStop)

	// Wait for cache to sync
	logger.Debug().Msg("waiting for cache sync...")
	if !cache.WaitForCacheSync(ctx.Done(), ctrl.HasSynced) {
		return fmt.Errorf("failed to sync Controller cache")
	}
	logger.Info().Msg("Controller cache synced")

	return nil
}

func (c *Controller) RunOrDie(ctx context.Context) {
	logger.Info().Msgf("starting %d workers", c.workers)

	// Start workers
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go func() {
			defer c.wg.Done()
			wait.UntilWithContext(
				ctx,
				func(ctx context.Context) {
					c.runWorker(ctx)
				}, time.Second)
		}()
	}

	// BLOCK until leadership is lost
	<-ctx.Done()

	logger.Info().Msg("leadership lost — draining workers...")

	// Stop accepting new items
	c.queue.ShutDown()

	// Wait for all workers to finish
	c.wg.Wait()

	logger.Info().Msg("controller drained and stopped")
}

// RegisterReconcilers registers all reconcilers to controller
func (c *Controller) RegisterReconcilers(r domain.Reconciler) {
	c.reconcilers = append(c.reconcilers, r)
	logger.Info().Msgf("%s reconciler egistered", r.Resource())
}

// RegisterReconcilers registers all reconcilers to controller
func (c *Controller) AddInformer(i *informer.Informer) {
	c.informers = append(c.informers, i)
	logger.Info().Msgf("%s informer added", i.Name())
}

// Shutdown gracefully stops the Controller
func (c *Controller) Shutdown(ctx context.Context) {
	logger.Info().Msg("shutting down Controller")
	c.queue.ShutDown()
}

// Controller name
func (c *Controller) Name() string {
	return "smart controller"
}
