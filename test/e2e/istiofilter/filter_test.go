package istiofilter

import (
	"testing"
	"time"

	"github.com/istio-conductor/istiofilter/test/e2e/utils"
)

func TestMain(m *testing.M) {
	utils.EnableVerbose()
	deployBookInfo()
	//defer clearBookInfo()
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
	mirror, err = utils.Kubectl().Get("VirtualService").ByName("productpage").ObjectField(".spec.http[0].mirror")
	if err != nil {
		t.Fatal(err)
	}
	wanted = ""
	if mirror != wanted {
		t.Fatalf("excepted mirror: %s but got: %s", wanted, mirror)
	}
}
