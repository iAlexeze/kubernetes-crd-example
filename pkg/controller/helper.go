package controller

import (
	"context"

	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/logger"
)

// runWorker is a long-running function that processes items from the queue
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextItem(ctx) {
	}
}

// processNextItem processes one item from the queue
func (c *Controller) processNextItem(ctx context.Context) bool {
	wq := c.wq.Queue

	// Wait until there's an item or the queue is shut down
	item, shutdown := wq.Get()
	if shutdown {
		return false
	}

	// We call Done at the end of this function to mark the item as processed
	defer wq.Done(item)

	// Find the right reconciler
	var targetReconciler domain.Reconciler
	for _, r := range c.reconcilers {
		if r.Resource().String() == item.Resource {
			targetReconciler = r
			break
		}
	}

	if targetReconciler == nil {
		logger.Error().Str("resource", item.Resource).Msg("no reconciler found")
		wq.Forget(item)
		return true
	}

	// Reconcile
	if err := targetReconciler.Reconcile(ctx, item.Key); err != nil {
		logger.Error().
			Err(err).
			Str("resource", item.Resource).
			Str("key", item.Key).
			Msg("reconcile failed")
		wq.AddRateLimited(item)
		return true
	}

	wq.Forget(item)
	return true
}
