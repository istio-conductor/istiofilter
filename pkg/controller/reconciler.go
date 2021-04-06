package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	api "github.com/istio-conductor/istiofilter/api/v1alpha1"
	"github.com/istio-conductor/istiofilter/client-go/pkg/apis/configuration/v1alpha1"
	"github.com/istio-conductor/istiofilter/pkg/annotation"
	"github.com/sirupsen/logrus"
	"istio.io/istio/pkg/config/schema/collections"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const finalizer = "finalizer.istiofilter.configuration.istio-conductor.org"
const finalizerMaxRetries = 2

type IstioFilterController struct {
	privilegeNamespaces map[string]struct{}
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	dn     dynamic.Interface
}

func NewIstioFilterController(privilegeNamespaces []string, client client.Client, dn dynamic.Interface) *IstioFilterController {
	ifc := &IstioFilterController{privilegeNamespaces: map[string]struct{}{}, client: client, dn: dn}
	for _, namespace := range privilegeNamespaces {
		ifc.privilegeNamespaces[namespace] = struct{}{}
	}
	return ifc
}

func (i *IstioFilterController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	ns, name := request.Namespace, request.Name
	logrus.Infof("Reconcile %s %s", ns, name)
	objKey := types.NamespacedName{
		Name:      request.Name,
		Namespace: ns,
	}
	var filter = &v1alpha1.IstioFilter{}
	err := i.client.Get(ctx, objKey, filter)
	if err != nil {
		if errors.IsNotFound(err) {
			logrus.Infof("Reconcile on not found resource: %s", request.NamespacedName)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	deleted := filter.GetDeletionTimestamp() != nil
	finalizers := sets.NewString(filter.GetFinalizers()...)
	var onSuccess func() (reconcile.Result, error)
	if deleted {
		logrus.Infof("Reconcile on delete resource: %s", request.NamespacedName)
		if !finalizers.Has(finalizer) {
			return reconcile.Result{}, nil
		}
		onSuccess = func() (reconcile.Result, error) {
			logrus.Infof("Reconcile empty resource finalizer: %s", request.NamespacedName)
			finalizers.Delete(finalizer)
			filter.SetFinalizers(finalizers.List())
			finalizerError := i.client.Update(context.TODO(), filter)
			for retryCount := 0; errors.IsConflict(finalizerError) && retryCount < finalizerMaxRetries; retryCount++ {
				_ = i.client.Get(context.TODO(), request.NamespacedName, filter)
				finalizers = sets.NewString(filter.GetFinalizers()...)
				finalizers.Delete(finalizer)
				filter.SetFinalizers(finalizers.List())
				finalizerError = i.client.Update(context.TODO(), filter)
			}
			if finalizerError != nil {
				if errors.IsNotFound(finalizerError) {
					return reconcile.Result{}, nil
				} else if errors.IsConflict(finalizerError) {
					logrus.Infof("Could not remove finalizer from %s due to conflict. Operation will be retried in next reconcile attempt.", filter.Name)
					return reconcile.Result{}, nil
				}
				logrus.Errorf("error removing finalizer: %s", finalizerError)
				return reconcile.Result{}, finalizerError
			}
			return reconcile.Result{}, nil
		}
	} else if !finalizers.Has(finalizer) {
		finalizers.Insert(finalizer)
		filter.Finalizers = finalizers.List()
		err := i.client.Update(context.TODO(), filter)
		if err != nil {
			if errors.IsNotFound(err) {
				logrus.Infof("Could not add finalizer to %s: the object was deleted.", filter.Name)
				return reconcile.Result{}, nil
			} else if errors.IsConflict(err) {
				logrus.Infof("Could not add finalizer to %s due to conflict. Operation will be retried in next reconcile attempt.", filter.Name)
				return reconcile.Result{}, nil
			}
			logrus.Errorf("Failed to add finalizer to IstioOperator CR %s: %s", filter.Name, err)
			return reconcile.Result{}, err
		}
	}

	logrus.Debug("spec:", filter.Spec)
	if filter.ResourceVersion == "" {
		return reconcile.Result{Requeue: true}, nil
	}

	rv, _ := strconv.ParseInt(filter.ResourceVersion, 10, 64)
	if deleted {
		rv = -1
	}
	var gvrs []schema.GroupVersionResource
	switch filter.Spec.Schema {
	case api.IstioFilter_DESTINATION_RULE:
		gvrs = append(gvrs, collections.IstioNetworkingV1Alpha3Destinationrules.Resource().GroupVersionResource())
	case api.IstioFilter_VIRTUAL_SERVICE:
		gvrs = append(gvrs, collections.IstioNetworkingV1Alpha3Virtualservices.Resource().GroupVersionResource())
	}

	for _, selector := range filter.Spec.Selectors {
		targetNs := selector.Namespace
		if targetNs == "" {
			targetNs = ns
		}
		if ns != targetNs {
			if _, ok := i.privilegeNamespaces[ns]; !ok {
				logrus.Warnf("[controller] ns %s not privilege", ns)
				continue
			}
		}
		options := SelectorToListOptions(selector)

		merge := AnnotationMerge{}
		merge.Metadata.Annotation = map[string]string{
			annotation.IstioFilterApplyName: fmt.Sprintf("%s/%s/%d", ns, name, rv),
		}
		for _, gvr := range gvrs {
			result, err := i.drive(ctx, driveContext{
				targetNs: targetNs,
				gvr:      gvr,
				options:  options,
				ns:       ns,
				name:     name,
				rv:       rv,
				patch:    []byte(merge.String()),
			})
			if err != nil {
				return reconcile.Result{}, err
			}
			if result.Requeue {
				return result, nil
			}
		}
	}
	if onSuccess != nil {
		return onSuccess()
	}
	return reconcile.Result{}, nil
}

type driveContext struct {
	targetNs string
	gvr      schema.GroupVersionResource

	options v1.ListOptions
	ns      string
	name    string
	rv      int64
	patch   []byte
}

func (i *IstioFilterController) drive(ctx context.Context, d driveContext) (reconcile.Result, error) {
	list, err := i.dn.Resource(d.gvr).Namespace(d.targetNs).List(ctx, d.options)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{Requeue: true}, nil
	}
	allApplied := true
	for _, item := range list.Items {
		applied := checkStatus(
			fmt.Sprintf("%s/%s/%s", item.GetKind(), item.GetNamespace(), item.GetName()),
			item.GetAnnotations()[annotation.IstioFilterStatusName],
			d.ns, d.name, d.rv)
		if applied {
			continue
		}
		allApplied = false
		logrus.Infof("%s\n", string(d.patch))
		_, err := i.dn.Resource(d.gvr).Namespace(d.targetNs).Patch(ctx, item.GetName(), types.MergePatchType, d.patch, v1.PatchOptions{})
		if err != nil {
			logrus.Error(err)
			return reconcile.Result{Requeue: true}, err
		}
	}
	if allApplied {
		return reconcile.Result{}, nil
	}
	return reconcile.Result{
		Requeue:      true,
		RequeueAfter: 1 * time.Second,
	}, nil
}

type AnnotationMerge struct {
	Metadata struct {
		Annotation map[string]string `json:"annotations"`
	} `json:"metadata"`
}

func (a *AnnotationMerge) String() string {
	bytes, _ := json.Marshal(a)
	return string(bytes)
}

func checkStatus(resourceName, str string, ns, n string, rv int64) (ok bool) {
	var l []string
	err := json.Unmarshal([]byte(str), &l)
	if err != nil {
		if rv <= 0 {
			return true
		}
		logrus.Warnf("[controller] resource %s status invalid", resourceName)
		return false
	}
	for _, s := range l {
		namespace, name, curr, err := annotation.Status(s)
		if err != nil {
			logrus.Warnf("[controller] resource %s status invalid %s", resourceName, s)
			continue
		}
		if namespace == ns && name == n {
			if rv <= 0 && curr > 0 {
				return false
			}
			return curr >= rv
		}
	}
	if rv <= 0 {
		return true
	}
	return false
}

func SelectorToListOptions(s *api.IstioFilter_Selector) v1.ListOptions {
	if s.Name != "" {
		return v1.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.name", s.Name).String()}
	}
	return v1.ListOptions{LabelSelector: labels.Set(s.LabelSelector).AsSelector().String()}
}
