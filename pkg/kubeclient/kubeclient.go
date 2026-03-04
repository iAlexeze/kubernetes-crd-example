package kubeclient

import (
	"context"

	"github.com/ialexeze/kubernetes-crd-example/pkg/config/api/types/v1alpha1"
	"github.com/ialexeze/kubernetes-crd-example/pkg/config/domain"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type kubeclient struct {
	isCustom   bool
	name       string
	restConfig *rest.Config
	clientset  kubernetes.Interface
	restClient rest.Interface

	Opts Options
}

type Options struct {
	Kubeconfig string
	Masterurl  string
}

var _ domain.Component = (*kubeclient)(nil)

func NewKubeclient(isCustom bool, opts Options) *kubeclient {
	return &kubeclient{
		name: "kubeclient",
		Opts: opts,
	}
}

func (k *kubeclient) Start(ctx context.Context) error {
	cfg, err := k.buildConfig()
	if err != nil {
		return err
	}

	// Populate kubeclient's rest config
	k.restConfig = cfg

	// Build clientset and restClient conditonally
	if k.isCustom {
		k.restClient, err = k.buildRestClient()
		if err != nil {
			return err
		}
	} else {
		k.clientset, err = kubernetes.NewForConfig(k.restConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (k *kubeclient) buildConfig() (*rest.Config, error) {
	if k.Opts.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(k.Opts.Masterurl, k.Opts.Kubeconfig)
	}
	return rest.InClusterConfig()
}

func (k *kubeclient) buildRestClient() (*rest.RESTClient, error) {
	config := *k.restConfig

	config.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   v1alpha1.GroupName,
		Version: v1alpha1.GroupVersion,
	}

	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.RESTClientFor(&config)
}

// Methods
func (k *kubeclient) Shutdown(ctx context.Context) {}

func (k *kubeclient) Name() string {
	return k.name
}

func (k *kubeclient) Clientset() kubernetes.Interface {
	return k.clientset
}

func (k *kubeclient) RestClient() rest.Interface {
	return k.restClient
}
