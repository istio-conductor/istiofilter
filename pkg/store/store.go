package store

import (
	"context"

	"github.com/istio-conductor/istiofilter/client-go/pkg/apis/istiofilter/v1alpha1"
)

type Store interface {
	Find(ctx context.Context, kind, name, namespace string, labels map[string]string) ([]*v1alpha1.IstioFilter, error)
}

type Mock struct {
	Filters []*v1alpha1.IstioFilter
}

func (m *Mock) Find(ctx context.Context, kind, name, namespace string, labels map[string]string) ([]*v1alpha1.IstioFilter, error) {
	return m.Filters, nil
}
