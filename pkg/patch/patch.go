package patch

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"strconv"

	"github.com/istio-conductor/istiofilter/api/v1alpha1"
	"github.com/istio-conductor/istiofilter/pkg/annotation"
	"github.com/istio-conductor/istiofilter/pkg/store"
	"github.com/istio-conductor/istiofilter/pkg/tricks"
	"github.com/sirupsen/logrus"
	"istio.io/istio/pilot/pkg/config/kube/crd"
	"istio.io/istio/pkg/config"
	"istio.io/istio/pkg/config/schema/collections"
	"istio.io/istio/pkg/config/schema/gvk"
	"k8s.io/apimachinery/pkg/util/json"
)

type Patcher struct {
	store store.Store
}

func New(store store.Store) *Patcher {
	return &Patcher{store: store}
}

var (
	ErrInvalidChanges = errors.New("invalid change")
	ErrCannotApply    = errors.New("cannot apply")
)

type statusApply struct {
	namespace string
	name      string
	rv        int64
}

func (p *Patcher) Filter(ctx context.Context, c *config.Config, oldCfg *config.Config) error {
	apply := c.Annotations[annotation.IstioFilterApplyName]
	if apply != "" {
		namespace, name, rv, err := annotation.Status(apply)
		if err != nil {
			return err
		}
		sa := &statusApply{namespace, name, rv}
		if oldCfg != nil && !reflect.DeepEqual(c.Spec, oldCfg.Spec) {
			return ErrInvalidChanges
		}
		var skeleton config.Spec
		skeletonConf := c.Annotations[annotation.IstioSkeletonName]
		if skeletonConf == "" {
			data, err := tricks.ToJSON(c.Spec)
			if err != nil {
				return err
			}
			c.Annotations[annotation.IstioSkeletonName] = string(data)
			skeleton = c.Spec
		} else {
			schema, _ := collections.Istio.FindByGroupVersionKind(c.GroupVersionKind)
			spec, err := crd.FromJSON(schema, skeletonConf)
			if err != nil {
				return err
			}
			skeleton = spec
		}
		c.Spec = skeleton
		err = p.PatchTo(ctx, c, sa)
		if err == nil {
			delete(c.Annotations, annotation.IstioFilterApplyName)
		}
		return err
	}
	if c.Annotations == nil {
		c.Annotations = map[string]string{}
	}

	data, _ := tricks.ToJSON(c.Spec)

	c.Annotations[annotation.IstioSkeletonName] = string(data)

	return p.PatchTo(ctx, c, nil)
}

func (p *Patcher) PatchTo(ctx context.Context, c *config.Config, apply *statusApply) error {
	fn, ok := ApplyTable[c.GroupVersionKind.Kind]
	if !ok {
		return nil
	}
	filters, err := p.store.Find(ctx, c.GroupVersionKind.Kind, c.Name, c.Namespace, c.Labels)
	if err != nil {
		return err
	}
	// 无selector的总是先执行
	// selector 为名字匹配的先执行，若都有，则匹配本身的名字长度
	sort.Slice(filters, func(i, j int) bool {
		fi := filters[i]
		fj := filters[j]
		if len(fi.Spec.Selectors) == 0 {
			return true
		}
		if len(fj.Spec.Selectors) == 0 {
			return true
		}
		if fi.Spec.Selectors[0].Name != "" && fj.Spec.Selectors[0].Name != "" {
			return len(fi.Name) < len(fj.Name)
		}
		if fi.Spec.Selectors[0].Name != "" {
			return false
		}
		if fj.Spec.Selectors[0].Name != "" {
			return false
		}
		return len(fi.Spec.Selectors[0].LabelSelector) < len(fj.Spec.Selectors[0].LabelSelector)
	})
	if apply != nil && apply.rv <= 0 {
		for _, filter := range filters {
			if filter.Namespace == apply.namespace && filter.Name == apply.name {
				return ErrCannotApply
			}
		}
	}
	var applied []string

	for _, filter := range filters {
		logrus.Infof("[patcher] patch %s to %s/%s", filter.Name, c.GroupVersionKind.Kind, c.Name)
		err := fn(c, &filter.Spec)
		if err != nil {
			logrus.Warnf("[patcher] %s/%s failed %s", filter.Namespace, filter.Name, err)
			continue
		}
		rv, _ := strconv.ParseInt(filter.ResourceVersion, 10, 64)
		status := annotation.ToStatus(filter.Namespace, filter.Name, rv)
		if apply != nil {
			if apply.namespace == filter.Namespace && apply.name == filter.Name {
				if rv < apply.rv {
					return ErrCannotApply
				}
			}
		}
		applied = append(applied, status)
	}

	appliedBytes, _ := json.Marshal(applied)
	c.Annotations[annotation.IstioFilterStatusName] = string(appliedBytes)
	return nil
}

var ApplyTable = map[string]func(c *config.Config, f *v1alpha1.IstioFilter) error{
	gvk.DestinationRule.Kind: ApplyToDestinationRule,
	gvk.VirtualService.Kind:  ApplyToVirtualService,
}
