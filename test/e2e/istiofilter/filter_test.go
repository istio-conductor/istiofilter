package istiofilter

import (
	"testing"
	"time"

	"github.com/istio-conductor/istiofilter/test/e2e/utils"
)

func TestMain(m *testing.M) {
	utils.EnableVerbose()
	deployBookInfo()
	defer clearBookInfo()
	m.Run()
}

func deployBookInfo() {
	_, err := utils.Kubectl().Apply("../bookinfo/bookinfo.yaml")
	if err != nil {
		panic(err)
	}
	_, err = utils.Kubectl().Apply("../bookinfo/destination-rule-all.yaml")
	if err != nil {
		panic(err)
	}
	_, err = utils.Kubectl().Apply("../bookinfo/virtual-service-all-v1.yaml")
	if err != nil {
		panic(err)
	}
	utils.Kubectl().WaitAllDeployments()
}

func clearBookInfo() {
	utils.Kubectl().DeleteByFile("../bookinfo/bookinfo.yaml")
	utils.Kubectl().DeleteByFile("../bookinfo/destination-rule-all.yaml")
	utils.Kubectl().DeleteByFile("../bookinfo/virtual-service-all-v1.yaml")
	utils.Kubectl().DeleteByFile("./testdata/mirror.yaml")
	utils.Kubectl().DeleteByFile("./testdata/route.yaml")
	utils.Kubectl().DeleteByFile("./testdata/outlier_detection.yaml")
}

func TestPatchVirtualService(t *testing.T) {
	utils.Kubectl().Apply("./testdata/mirror.yaml")
	time.Sleep(time.Millisecond * 100)
	mirror, err := utils.Kubectl().Get("VirtualService").ByName("productpage").ObjectField(".spec.http[0].mirror")
	if err != nil {
		t.Fatal(err)
	}
	wanted := "{\"host\":\"productpage\",\"subset\":\"v2\"}"
	if mirror != wanted {
		t.Fatalf("excepted mirror: %s but got: %s", wanted, mirror)
	}
	utils.Kubectl().DeleteByFile("./testdata/mirror.yaml")
	time.Sleep(time.Millisecond * 100)
	mirror, err = utils.Kubectl().Get("VirtualService").ByName("productpage").ObjectField(".spec.http[0].mirror")
	if err != nil {
		t.Fatal(err)
	}
	wanted = ""
	if mirror != wanted {
		t.Fatalf("excepted mirror: %s but got: %s", wanted, mirror)
	}
	utils.Kubectl().Apply("./testdata/route.yaml")
	time.Sleep(time.Millisecond * 100)
	routeName,err:=utils.Kubectl().Get("VirtualService").ByName("productpage").ObjectField(".spec.http[0].name")
	wanted = "route"
	if routeName!=wanted{
		t.Fatalf("excepted first route name: %s but got: %s", wanted, routeName)
	}

	routeName,err=utils.Kubectl().Get("VirtualService").ByName("productpage").ObjectField(".spec.http[1].name")
	wanted = "route2"
	if routeName!=wanted{
		t.Fatalf("excepted first route name: %s but got: %s", wanted, routeName)
	}

	routeName,err=utils.Kubectl().Get("VirtualService").ByName("productpage").ObjectField(".spec.http[2].name")
	wanted = ""
	if routeName!=wanted{
		t.Fatalf("excepted first route name: %s but got: %s", wanted, routeName)
	}
}

func TestPatchDestination(t *testing.T) {
	utils.Kubectl().Apply("./testdata/outlier_detection.yaml")
	time.Sleep(time.Millisecond * 100)
	mirror, err := utils.Kubectl().Get("DestinationRule").ByName("productpage").ObjectField(".spec.trafficPolicy.outlierDetection")
	if err != nil {
		t.Fatal(err)
	}
	wanted := "{\"baseEjectionTime\":\"180s\",\"consecutiveErrors\":1,\"interval\":\"1s\",\"maxEjectionPercent\":100}"
	if mirror != wanted {
		t.Fatalf("excepted mirror: %s but got: %s", wanted, mirror)
	}
	utils.Kubectl().Get("DestinationRule").ByName("reviews").YAML()

	mirror, err = utils.Kubectl().Get("DestinationRule").ByName("reviews").ObjectField(".spec.subsets[0].trafficPolicy.outlierDetection")
	if err != nil {
		t.Fatal(err)
	}
	wanted = "{\"baseEjectionTime\":\"180s\",\"consecutiveErrors\":1,\"interval\":\"1s\",\"maxEjectionPercent\":100}"
	if mirror != wanted {
		t.Fatalf("excepted mirror: %s but got: %s", wanted, mirror)
	}

	utils.Kubectl().DeleteByFile("./testdata/outlier_detection.yaml")
	mirror, err = utils.Kubectl().Get("DestinationRule").ByName("productpage").ObjectField(".spec.trafficPolicy.outlierDetection")
	if err != nil {
		t.Fatal(err)
	}
	wanted = ""
	if mirror != wanted {
		t.Fatalf("excepted mirror: %s but got: %s", wanted, mirror)
	}
}