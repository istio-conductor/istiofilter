package webhook

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/istio-conductor/istiofilter/pkg/patch"
	"github.com/istio-conductor/istiofilter/pkg/tricks"
	"github.com/sirupsen/logrus"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/json"

	"istio.io/istio/pkg/config/schema/collections"

	kh "github.com/slok/kubewebhook/v2/pkg/http"
	ll "github.com/slok/kubewebhook/v2/pkg/log/logrus"
	"github.com/slok/kubewebhook/v2/pkg/model"
	"github.com/slok/kubewebhook/v2/pkg/webhook/mutating"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Config struct {
}

// Run webhook
func Run(ctx context.Context, port int, certFile string, keyFile string, patcher *patch.Patcher) error {
	logger := ll.NewLogrus(logrus.NewEntry(logrus.New()))
	wh, _ := mutating.NewWebhook(mutating.WebhookConfig{
		ID:      "conductor",
		Logger:  logger,
		Mutator: &IstioMutator{patcher: patcher},
	})
	handler, err := kh.HandlerFor(kh.HandlerConfig{
		Webhook: wh,
		Logger:  logger,
	})
	if err != nil {
		return err
	}
	http.Handle("/mutate", handler)
	s := &http.Server{Addr: fmt.Sprintf(":%d", port)}

	go func() {
		<-ctx.Done()
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		_ = s.Shutdown(timeout)
	}()
	return s.ListenAndServeTLS(certFile, keyFile)
}

type IstioMutator struct {
	patcher *patch.Patcher
}

func jsonStr(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}

func (i *IstioMutator) Mutate(ctx context.Context, ar *model.AdmissionReview, object v1.Object) (result *mutating.MutatorResult, err error) {
	un, ok := object.(*unstructured.Unstructured)
	if !ok {
		return &mutating.MutatorResult{}, nil
	}
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debug("entry mutate:", jsonStr(un.Object))
	}
	g := un.GroupVersionKind()
	schema, exists := collections.PilotServiceApi.FindByGroupVersionKind(resource.FromKubernetesGVK(&g))
	if !exists {
		return &mutating.MutatorResult{}, nil
	}
	var oldCfg *config.Config
	if ar.Operation == model.OperationUpdate {
		ik, err := jsonToIstioKind(ar.OldObjectRaw)
		if err != nil {
			return nil, err
		}
		oldCfg, err = crd.ConvertObject(schema, ik, "")
		if err != nil {
			return nil, err
		}
	}

	cfg, err := crd.ConvertObject(schema, toIstioKind(un), "")
	if err != nil {
		return nil, err
	}

	err = i.patcher.Filter(ctx, cfg, oldCfg)
	if err != nil {
		return nil, err
	}

	out, err := toKubeObject(cfg)
	if err != nil {
		return nil, err
	}
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debug("after mutate:", jsonStr(out.Object))
	}
	return &mutating.MutatorResult{
		MutatedObject: out,
	}, nil
}

func mustGetMap(object map[string]interface{}, fields ...string) map[string]interface{} {
	m, _, _ := unstructured.NestedMap(object, fields...)
	return m
}

func jsonToIstioKind(data []byte) (*crd.IstioKind, error) {
	ik := &crd.IstioKind{}
	err := json.Unmarshal(data, ik)
	if err != nil {
		return nil, err
	}
	return ik, nil
}

func toIstioKind(un *unstructured.Unstructured) *crd.IstioKind {
	return &crd.IstioKind{
		TypeMeta: metav1.TypeMeta{
			Kind:       un.GetKind(),
			APIVersion: un.GetAPIVersion(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:              un.GetName(),
			Namespace:         un.GetNamespace(),
			ResourceVersion:   un.GetResourceVersion(),
			Labels:            un.GetLabels(),
			Annotations:       un.GetAnnotations(),
			CreationTimestamp: un.GetCreationTimestamp(),
		},
		Spec:   mustGetMap(un.Object, "spec"),
		Status: mustGetMap(un.Object, "status"),
	}
}

func toMap(spec config.Spec) (map[string]interface{}, error) {
	fields := tricks.FindInt64Field(spec)
	m, err := config.ToMap(spec)
	if err != nil {
		return nil, err
	}
	err = tricks.RestoreIntField(m, fields)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func toKubeObject(cfg *config.Config) (*unstructured.Unstructured, error) {
	spec, err := toMap(cfg.Spec)
	if err != nil {
		return nil, err
	}
	status, err := toMap(cfg.Status)
	if err != nil {
		return nil, err
	}

	namespace := cfg.Namespace
	if namespace == "" {
		namespace = metav1.NamespaceDefault
	}
	un := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec":   spec,
			"status": status,
		},
	}
	un.SetAPIVersion(cfg.GroupVersionKind.GroupVersion())
	un.SetKind(cfg.GroupVersionKind.Kind)
	un.SetName(cfg.Name)
	un.SetNamespace(namespace)
	un.SetResourceVersion(cfg.ResourceVersion)
	un.SetLabels(cfg.Labels)
	un.SetAnnotations(cfg.Annotations)
	un.SetCreationTimestamp(metav1.NewTime(cfg.CreationTimestamp))
	return un, nil
}
