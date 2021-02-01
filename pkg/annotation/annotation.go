package annotation

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/json"
)

const IstioFilterApplyName = "istio-conductor.org/istio-filter-apply"
const IstioFilterStatusName = "istio-conductor.org/istio-filter-status"
const IstioSkeletonName = "istio-conductor.org/skeleton"
const IstioSpecHash = "istio-conductor.org/spec-hash"

var (
	ErrInvalidStatus = errors.New("invalid status")
)

func ToStatus(namespace string, name string, rv int64) string {
	return fmt.Sprintf("%s/%s/%d", namespace, name, rv)
}

type FilterID struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ResourceVersion int64  `json:"rv"`
}

func Statuses(status string) (filters map[string]FilterID) {
	filters = map[string]FilterID{}
	var ss []string
	_ = json.Unmarshal([]byte(status), &ss)
	for _, s := range ss {
		namespace, name, rv, err := Status(s)
		if err != nil {
			continue
		}
		filters[namespace+"/"+name] = FilterID{
			namespace, name, rv,
		}
	}
	return filters
}

func Status(status string) (namespace string, name string, rv int64, err error) {
	nnr := strings.Split(status, "/")
	if len(nnr) != 3 {
		err = ErrInvalidStatus
		return
	}
	rv, err = strconv.ParseInt(nnr[2], 10, 64)
	if err != nil {
		err = ErrInvalidStatus
		return
	}
	namespace = nnr[0]
	name = nnr[1]
	return
}
