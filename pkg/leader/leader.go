package leader

import (
	"context"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/pkg/logger"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type leaderElection struct {
	name string
	kube *kubeclient.Kubeclient
	run  func(context.Context)

	opts Options
}

type Options struct {
	LeaseDuration time.Duration
	RetryPeriod   time.Duration
	RenewDeadline time.Duration

	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

var _ domain.Component = (*leaderElection)(nil)

func NewLeaderElection(kube *kubeclient.Kubeclient, run func(context.Context), opts Options) *leaderElection {
	if opts.Namespace == "" {
		opts.Namespace = "default"
	}

	return &leaderElection{
		name: "resource-leader",
		kube: kube,
		opts: opts,
	}
}

func (le *leaderElection) Start(ctx context.Context) error {

	go func() {
		leaderelection.RunOrDie(ctx, le.leaseConfig())
	}()
	return nil
}

func (le *leaderElection) Shutdown(ctx context.Context) {}

func (le *leaderElection) Name() string {
	return le.name
}

// Helpers
// Lease configuration
func (le *leaderElection) leaseConfig() leaderelection.LeaderElectionConfig {
	return leaderelection.LeaderElectionConfig{
		Name:            le.Name(),
		Lock:            le.leaseLock(),
		LeaseDuration:   le.opts.LeaseDuration,
		RenewDeadline:   le.opts.RenewDeadline,
		RetryPeriod:     le.opts.RetryPeriod,
		ReleaseOnCancel: true,
		Callbacks:       le.callbacks(),
	}
}

// Lease lock
func (le *leaderElection) leaseLock() *resourcelock.LeaseLock {
	opts := le.opts
	return &resourcelock.LeaseLock{
		LeaseMeta: v1.ObjectMeta{
			Name:        le.name,
			Namespace:   opts.Namespace,
			Annotations: opts.Annotations,
			Labels:      opts.Labels,
		},
		Client: le.kube.Clientset().CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      hostname(),
			EventRecorder: nil,
		},
	}
}

// Build callbacks
func (le *leaderElection) callbacks() leaderelection.LeaderCallbacks {
	return leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) { le.run(ctx) },
		OnStoppedLeading: func() { logger.Info().Msg("leader lost") },
	}
}

// Get hostname
func hostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = uuid.New().String()
	}
	return hostname
}
