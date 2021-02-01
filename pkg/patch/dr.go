package patch

import (
	"github.com/istio-conductor/istiofilter/api/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
)

func ApplyToDestinationRule(cfg *config.Config, filter *v1alpha1.IstioFilter) (err error) {
	rule := cfg.Spec.(*networking.DestinationRule)
	for _, change := range filter.Changes {
		err = applyToDestinationRuleChange(rule, change)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyToDestinationRuleChange(rule *networking.DestinationRule, change *v1alpha1.IstioFilter_Change) (err error) {
	switch change.ApplyTo {
	case v1alpha1.IstioFilter_LOAD_BALANCER:
		changeTraffic(rule, change, applyToLoadBalancer)
		return
	case v1alpha1.IstioFilter_OUTLIER_DETECTION:
		changeTraffic(rule, change, applyToOutlierDetection)
		return
	case v1alpha1.IstioFilter_CONNECTION_POOL:
		changeTraffic(rule, change, applyToConnectionPool)
		return
	}
	return ErrUnknownContext
}

func changeTraffic(rule *networking.DestinationRule, change *v1alpha1.IstioFilter_Change, apply applyChangeToTraffic) {
	match := change.Match.Match
	switch cvt := match.(type) {
	case *v1alpha1.IstioFilter_Match_Simple:
		switch cvt.Simple {
		case v1alpha1.IstioFilter_ALL:
			rule.TrafficPolicy = apply(rule.TrafficPolicy, change.Patch)
			for _, subset := range rule.Subsets {
				subset.TrafficPolicy = apply(subset.TrafficPolicy, change.Patch)
			}
		case v1alpha1.IstioFilter_DEFAULT:
			rule.TrafficPolicy = apply(rule.TrafficPolicy, change.Patch)
		}
	case *v1alpha1.IstioFilter_Match_Selector:
		if cvt.Selector.Name != nil {
			for _, subset := range rule.Subsets {
				if !cvt.Selector.Name.MatchValue(subset.Name) {
					continue
				}
				subset.TrafficPolicy = apply(subset.TrafficPolicy, change.Patch)
			}
		}
		if cvt.Selector.Labels != nil {
			for _, subset := range rule.Subsets {
				if !matchLabels(subset.Labels, cvt.Selector.Labels) {
					continue
				}
				subset.TrafficPolicy = apply(subset.TrafficPolicy, change.Patch)
			}
		}
	}
}

type applyChangeToTraffic func(traffic *networking.TrafficPolicy, patch *v1alpha1.IstioFilter_Patch) *networking.TrafficPolicy

func applyToLoadBalancer(traffic *networking.TrafficPolicy, patch *v1alpha1.IstioFilter_Patch) *networking.TrafficPolicy {
	if traffic == nil {
		if patch.Operation == v1alpha1.IstioFilter_REMOVE {
			return nil
		}
		traffic = &networking.TrafficPolicy{
			LoadBalancer: &networking.LoadBalancerSettings{},
		}
	}
	lb := &networking.LoadBalancerSettings{}
	switch patch.Operation {
	case v1alpha1.IstioFilter_MERGE:
		if traffic.LoadBalancer != nil {
			lb = traffic.LoadBalancer.DeepCopy()
		}
		merge(lb, patch.Value)
	case v1alpha1.IstioFilter_REPLACE:
		merge(lb, patch.Value)
	case v1alpha1.IstioFilter_REMOVE:
		lb = nil
	}
	traffic.LoadBalancer = lb
	return traffic
}

func applyToOutlierDetection(traffic *networking.TrafficPolicy, patch *v1alpha1.IstioFilter_Patch) *networking.TrafficPolicy {
	if traffic == nil {
		if patch.Operation == v1alpha1.IstioFilter_REMOVE {
			return nil
		}
		traffic = &networking.TrafficPolicy{
			OutlierDetection: &networking.OutlierDetection{},
		}
	}
	od := &networking.OutlierDetection{}
	switch patch.Operation {
	case v1alpha1.IstioFilter_MERGE:
		if traffic.OutlierDetection != nil {
			od = traffic.OutlierDetection.DeepCopy()
		}
		merge(od, patch.Value)
	case v1alpha1.IstioFilter_REPLACE:
		merge(od, patch.Value)
	case v1alpha1.IstioFilter_REMOVE:
		od = nil
	}
	traffic.OutlierDetection = od
	return traffic
}

func applyToConnectionPool(traffic *networking.TrafficPolicy, patch *v1alpha1.IstioFilter_Patch) *networking.TrafficPolicy {
	if traffic == nil {
		if patch.Operation == v1alpha1.IstioFilter_REMOVE {
			return nil
		}
		traffic = &networking.TrafficPolicy{
			ConnectionPool: &networking.ConnectionPoolSettings{},
		}
	}
	od := &networking.ConnectionPoolSettings{}
	switch patch.Operation {
	case v1alpha1.IstioFilter_MERGE:
		if traffic.OutlierDetection != nil {
			od = traffic.ConnectionPool.DeepCopy()
		}
		merge(od, patch.Value)
	case v1alpha1.IstioFilter_REPLACE:
		merge(od, patch.Value)
	case v1alpha1.IstioFilter_REMOVE:
		od = nil
	}
	traffic.ConnectionPool = od
	return traffic
}
