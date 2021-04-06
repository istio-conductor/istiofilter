package patch

import (
	"bytes"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/istio-conductor/istiofilter/api/v1alpha1"
	networking "istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
)

func ApplyToVirtualService(cfg *config.Config, filter *v1alpha1.IstioFilter) (err error) {
	rule := cfg.Spec.(*networking.VirtualService)
	for _, change := range filter.Changes {
		err = applyToVirtualServiceChange(rule, change)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyToVirtualServiceChange(vs *networking.VirtualService, change *v1alpha1.IstioFilter_Change) (err error) {
	// prepare value
	switch change.ApplyTo {
	case v1alpha1.IstioFilter_HTTP_ROUTE:
		return changeRoute(vs, change, applyToHTTPRoute)
	case v1alpha1.IstioFilter_HTTP_ROUTE_FAULT:
		return changeRoute(vs, change, applyToHTTPRouteFault)
	}
	return ErrUnknownContext
}

func changeRoute(vs *networking.VirtualService, change *v1alpha1.IstioFilter_Change, apply func(route *networking.HTTPRoute, patch *v1alpha1.IstioFilter_Patch) (*networking.HTTPRoute, error)) (err error) {
	var defaultRoute *networking.HTTPRoute
	var defaultIndex int
	for i, route := range vs.Http {
		if len(route.Match) == 0 {
			defaultRoute = route
			defaultIndex = i
		}
	}
	switch change.Match.GetMatch().(type) {
	case *v1alpha1.IstioFilter_Match_Simple:
		switch change.Match.GetSimple() {
		case v1alpha1.IstioFilter_ALL:
			switch change.Patch.Operation {
			case v1alpha1.IstioFilter_REMOVE,
				v1alpha1.IstioFilter_REPLACE,
				v1alpha1.IstioFilter_INSERT_BEFORE,
				v1alpha1.IstioFilter_INSERT_AFTER:
				return ErrUnsupportedOperation
			}
			// only support merge
			for i, route := range vs.Http {
				vs.Http[i], err = apply(route, change.Patch)
				if err != nil {
					return err
				}
			}
		case v1alpha1.IstioFilter_DEFAULT:
			if change.Patch.Operation == v1alpha1.IstioFilter_REMOVE {
				return ErrUnsupportedOperation
			}
			if defaultRoute == nil {
				return nil
			}
			applied, err := apply(defaultRoute, change.Patch)
			if err != nil {
				return err
			}
			switch change.Patch.Operation {
			case v1alpha1.IstioFilter_INSERT_BEFORE:
				temp := append(make([]*networking.HTTPRoute, 0, len(vs.Http)+1), vs.Http[:defaultIndex]...)
				temp = append(temp, applied)
				temp = append(temp, vs.Http[defaultIndex:]...)
				vs.Http = temp
			case v1alpha1.IstioFilter_INSERT_AFTER:
				temp := append(make([]*networking.HTTPRoute, 0, len(vs.Http)+1), vs.Http[:defaultIndex+1]...)
				temp = append(temp, applied)
				temp = append(temp, vs.Http[defaultIndex+1:]...)
				vs.Http = temp
			case v1alpha1.IstioFilter_MERGE, v1alpha1.IstioFilter_REPLACE:
				vs.Http[defaultIndex] = applied
			}
		}
	case *v1alpha1.IstioFilter_Match_Selector:
		if change.Match.GetSelector().Name == nil {
			return nil
		}
		temp := make([]*networking.HTTPRoute, 0, len(vs.Http))
		for _, route := range vs.Http {
			if !change.Match.GetSelector().Name.MatchValue(route.Name) {
				temp = append(temp, route)
				continue
			}
			applied, err := apply(route, change.Patch)
			if err != nil {
				return err
			}
			if applied != nil {
				switch change.Patch.Operation {
				case v1alpha1.IstioFilter_INSERT_BEFORE:
					temp = append(temp, applied, route)
				case v1alpha1.IstioFilter_INSERT_AFTER:
					temp = append(temp, route, applied)
				case v1alpha1.IstioFilter_MERGE:
					temp = append(temp, applied)
					if route.Name != applied.Name {
						temp = append(temp, route)
					}
				case v1alpha1.IstioFilter_REPLACE:
					temp = append(temp, applied)
				}
			}
		}
		vs.Http = temp
	}
	return nil
}

func struct2Route(src *types.Struct) *networking.HTTPRoute {
	srcJS, _ := (&jsonpb.Marshaler{}).MarshalToString(src)
	v := &networking.HTTPRoute{}
	_ = (&jsonpb.Unmarshaler{AllowUnknownFields: true}).Unmarshal(bytes.NewBufferString(srcJS), v)
	return v
}

func mergeRoute(route *networking.HTTPRoute, src *types.Struct) *networking.HTTPRoute {
	cloned := route.DeepCopy()
	patchRoute := struct2Route(src)
	if len(patchRoute.Route) != 0 {
		cloned.Route = nil
	}
	if len(patchRoute.Match) != 0 {
		cloned.Match = nil
	}
	proto.Merge(cloned, patchRoute)
	return cloned
}

func applyToHTTPRoute(route *networking.HTTPRoute, patch *v1alpha1.IstioFilter_Patch) (*networking.HTTPRoute, error) {
	switch patch.Operation {
	case v1alpha1.IstioFilter_MERGE:
		return mergeRoute(route, patch.Value), nil
	case v1alpha1.IstioFilter_REPLACE,
		v1alpha1.IstioFilter_INSERT_AFTER,
		v1alpha1.IstioFilter_INSERT_BEFORE:
		cp := &networking.HTTPRoute{}
		merge(cp, patch.Value)
		return cp, nil
	case v1alpha1.IstioFilter_REMOVE:
		return nil, nil
	default:
		return nil, ErrUnsupportedOperation
	}
}

func applyToHTTPRouteFault(route *networking.HTTPRoute, patch *v1alpha1.IstioFilter_Patch) (*networking.HTTPRoute, error) {
	switch patch.Operation {
	case v1alpha1.IstioFilter_MERGE:
		if route.Fault == nil {
			route.Fault = &networking.HTTPFaultInjection{}
		}
		merge(route.Fault, patch.Value)
		return route, nil
	case v1alpha1.IstioFilter_REPLACE:
		cp := &networking.HTTPFaultInjection{}
		merge(cp, patch.Value)
		route.Fault = cp
		return route, nil
	case v1alpha1.IstioFilter_REMOVE:
		route.Fault = nil
		return route, nil
	default:
		return nil, ErrUnsupportedOperation
	}
}
