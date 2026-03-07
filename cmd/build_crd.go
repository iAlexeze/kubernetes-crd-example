package main

import (
	"reflect"

	managednsTypeV1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/managedNamespace/v1alpha1"
	projectTypev1 "github.com/ialexeze/multi-crd-controller/pkg/config/api/types/project/v1alpha1"
	"github.com/ialexeze/multi-crd-controller/pkg/config/domain"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/event"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/kubeclient"
	"github.com/ialexeze/multi-crd-controller/pkg/config/pkg/reconciler"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type crd struct {
	obj        runtime.Object
	listObj    runtime.Object
	reconciler reconciler.NewReconcilerFunc
	info       kubeclient.CRDInfo
}

var ResourceTypeMap = map[reflect.Type]string{}

// Define new CRD
func NewCRD(
	obj runtime.Object,
	listObj runtime.Object,
	info kubeclient.CRDInfo,
	newRec reconciler.NewReconcilerFunc,
) crd {
	return crd{
		obj:        obj,
		listObj:    listObj,
		info:       info,
		reconciler: newRec,
	}
}

func buildCRDs(kube *kubeclient.Kubeclient) []crd {
	return []crd{
		NewCRD(
			&projectTypev1.Project{},
			&projectTypev1.ProjectList{},
			kubeclient.CRDInfoFrom(
				projectTypev1.Group,
				projectTypev1.Version,
				projectTypev1.Kind,
				projectTypev1.APIPath,
				projectTypev1.NamePlural,
				false,
			),
			func(inf cache.SharedIndexInformer, ev *event.Event) domain.Reconciler {
				return reconciler.NewProjectReconciler(inf, ev)
			},
		),
		NewCRD(
			&managednsTypeV1.ManagedNamespace{},
			&managednsTypeV1.ManagedNamespaceList{},
			kubeclient.CRDInfoFrom(
				managednsTypeV1.Group,
				managednsTypeV1.Version,
				managednsTypeV1.Kind,
				managednsTypeV1.APIPath,
				managednsTypeV1.NamePlural,
				true,
			),
			func(inf cache.SharedIndexInformer, ev *event.Event) domain.Reconciler {
				return reconciler.NewManagedNamespaceReconciler(kube, inf, ev)
			},
		),
	}
}
