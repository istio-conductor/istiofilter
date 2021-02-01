package controller

import (
	"context"
	"reflect"

	"github.com/istio-conductor/istiofilter/client-go/pkg/apis/istiofilter/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func AddIstioFilter(mgr manager.Manager, privilegeNamespaces []string) error {
	dn, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}
	go func() {
		watch, _ := dn.Resource(schema.GroupVersionResource{
			Group:    "istio-conductor.org",
			Version:  "v1alpha1",
			Resource: "istiofilters",
		}).Namespace("go-grpc-test-core").Watch(context.Background(), metav1.ListOptions{})
		for e := range watch.ResultChan() {
			un, ok := e.Object.(*unstructured.Unstructured)
			if ok {
				logrus.Info(un.Object["spec"])
			}
		}
	}()

	filterController := NewIstioFilterController(privilegeNamespaces, mgr.GetClient(), dn)
	c, err := controller.New("istio-conductor-filter-controller", mgr, controller.Options{Reconciler: filterController})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource IstioFilter
	err = c.Watch(&source.Kind{Type: &v1alpha1.IstioFilter{}}, &handler.EnqueueRequestForObject{}, filterPredicates)
	if err != nil {
		return err
	}
	return nil
}

var (
	filterPredicates = predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldFilter, ok := e.ObjectOld.(*v1alpha1.IstioFilter)
			if !ok {
				return false
			}
			newFilter := e.ObjectNew.(*v1alpha1.IstioFilter)
			if !ok {
				return false
			}
			if !reflect.DeepEqual(oldFilter.Spec, newFilter.Spec) ||
				oldFilter.GetDeletionTimestamp() != newFilter.GetDeletionTimestamp() ||
				oldFilter.GetGeneration() != newFilter.GetGeneration() {
				return true
			}
			return false
		},
	}
)
