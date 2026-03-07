package main

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	mnsTypev1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/managedNamespace/v1alpha1"
	projectTypev1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

func buildScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()

	// 1. Register built-in Kubernetes types
	metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)

	// 2. Register core Kubernetes types
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	// 3. Register your CRDs
	if err := projectTypev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := mnsTypev1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}
