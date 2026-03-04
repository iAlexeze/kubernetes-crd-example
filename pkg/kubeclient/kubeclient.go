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

type Kubeclient struct {
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

var _ domain.Component = (*Kubeclient)(nil)

func NewKubeclient(isCustom bool, opts Options) *Kubeclient {
	return &Kubeclient{
		name: "kubeclient",
		Opts: opts,
	}
}

func (k *Kubeclient) Start(ctx context.Context) error {
	cfg, err := k.buildConfig()
	if err != nil {
		return err
	}

	// Populate Kubeclient's rest config
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

func (k *Kubeclient) buildConfig() (*rest.Config, error) {
	if k.Opts.Kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags(k.Opts.Masterurl, k.Opts.Kubeconfig)
	}
	return rest.InClusterConfig()
}

func (k *Kubeclient) buildRestClient() (*rest.RESTClient, error) {
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
func (k *Kubeclient) Shutdown(ctx context.Context) {}

func (k *Kubeclient) Name() string {
	return k.name
}

func (k *Kubeclient) Clientset() kubernetes.Interface {
	return k.clientset
}

func (k *Kubeclient) RestClient() rest.Interface {
	return k.restClient
}
