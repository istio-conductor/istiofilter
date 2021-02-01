package tricks

import (
	"encoding/json"
	"testing"

	"istio.io/api/networking/v1alpha3"
	"istio.io/istio/pkg/config"
)

func TestFindInt64Field(t *testing.T) {
	spec := &v1alpha3.DestinationRule{
		TrafficPolicy: &v1alpha3.TrafficPolicy{
			LoadBalancer: &v1alpha3.LoadBalancerSettings{
				LbPolicy: &v1alpha3.LoadBalancerSettings_ConsistentHash{ConsistentHash: &v1alpha3.LoadBalancerSettings_ConsistentHashLB{
					MinimumRingSize: 32,
				}}},
		},
		Subsets: []*v1alpha3.Subset{{
			Name: "",
			TrafficPolicy: &v1alpha3.TrafficPolicy{
				LoadBalancer: &v1alpha3.LoadBalancerSettings{
					LbPolicy: &v1alpha3.LoadBalancerSettings_ConsistentHash{ConsistentHash: &v1alpha3.LoadBalancerSettings_ConsistentHashLB{
						MinimumRingSize: 64,
					}}},
			},
		},
		},
	}
	field := FindInt64Field(spec)

	m, _ := config.ToMap(spec)
	data, err := json.Marshal(m)
	t.Log(string(data), err)
	t.Log(field)
	RestoreIntField(m, field)

	data, err = json.Marshal(m)
	t.Log(string(data), err)

}
