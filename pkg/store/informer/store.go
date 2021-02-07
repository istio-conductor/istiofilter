package informer

import (
	"context"
	"github.com/istio-conductor/istiofilter/api/v1alpha1"
	k8sv1alpha1 "github.com/istio-conductor/istiofilter/client-go/pkg/apis/configuration/v1alpha1"
	"github.com/istio-conductor/istiofilter/client-go/pkg/clientset/versioned"
	"github.com/istio-conductor/istiofilter/client-go/pkg/informers/externalversions"
	listerv1 "github.com/istio-conductor/istiofilter/client-go/pkg/listers/configuration/v1alpha1"
	"istio.io/istio/pkg/config/labels"
	"istio.io/istio/pkg/config/schema/collections"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"time"
)

type Store struct {
	privilegeNamespaces map[string]struct{}
	factory             externalversions.SharedInformerFactory
	lister              listerv1.IstioFilterLister
	cli                 *versioned.Clientset
}

func New(privilegeNamespaces []string, cfg *rest.Config) (*Store, error) {
	s := &Store{privilegeNamespaces: map[string]struct{}{}}
	for _, namespace := range privilegeNamespaces {
		s.privilegeNamespaces[namespace] = struct{}{}
	}
	cli, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	s.cli = cli
	return s, nil
}

func (s *Store) Start(stop <-chan struct{}) {
	factory := externalversions.NewSharedInformerFactory(s.cli, time.Minute)
	s.factory = factory
	s.lister = factory.Configuration().V1alpha1().IstioFilters().Lister()
	factory.Start(stop)
	factory.WaitForCacheSync(stop)
}

func (s *Store) find(kind, name, namespace string, foundNamespace string, lbs map[string]string) (matched []*k8sv1alpha1.IstioFilter, err error) {
	filters, err := s.lister.IstioFilters(foundNamespace).List(klabels.Everything())
	if err != nil {
		return nil, err
	}
	for _, filter := range filters {
		if filter.DeletionTimestamp != nil {
			continue
		}
		SchemaKind, ok := schemeToKind[filter.Spec.Schema]
		if !ok {
			continue
		}
		if SchemaKind != kind {
			continue
		}

		match := false
		for _, selector := range filter.Spec.Selectors {
			if s.MatchFind(selector, name, namespace, lbs) {
				filter.Spec.Selectors = []*v1alpha1.IstioFilter_Selector{selector}
				match = true
				break
			}
		}
		if !match {
			continue
		}
		matched = append(matched, filter)
	}
	return matched, nil
}

func (s *Store) Find(ctx context.Context, kind, name, namespace string, labels map[string]string) (matched []*k8sv1alpha1.IstioFilter, err error) {
	matched, err = s.find(kind, name, namespace, namespace, labels)
	if err != nil {
		return nil, err
	}
	for ns := range s.privilegeNamespaces {
		privilegeMatched, err := s.find(kind, name, namespace, ns, labels)
		if err != nil {
			return nil, err
		}
		matched = append(matched, privilegeMatched...)
	}
	return removeDuplicatedFilters(matched), nil
}

func removeDuplicatedFilters(filters []*k8sv1alpha1.IstioFilter) (result []*k8sv1alpha1.IstioFilter) {
	set := map[string]*k8sv1alpha1.IstioFilter{}
	for _, filter := range filters {
		set[filter.Namespace+"/"+filter.Name] = filter
	}
	for _, filter := range set {
		result = append(result, filter)
	}
	return result
}

var schemeToKind = map[v1alpha1.IstioFilter_Schema]string{
	v1alpha1.IstioFilter_VIRTUAL_SERVICE:  collections.IstioNetworkingV1Alpha3Virtualservices.Resource().Kind(),
	v1alpha1.IstioFilter_DESTINATION_RULE: collections.IstioNetworkingV1Alpha3Destinationrules.Resource().Kind(),
}

func (s *Store) MatchFind(selector *v1alpha1.IstioFilter_Selector, name, namespace string, lbs map[string]string) bool {
	if selector.Namespace != "" && namespace != selector.Namespace {
		return false
	}
	if selector.Name != "" && name != selector.Name {
		return false
	}
	l := labels.Instance(selector.LabelSelector)
	return l.SubsetOf(lbs)
}
