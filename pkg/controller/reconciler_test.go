package controller

import "testing"

func TestCheckStatus(t *testing.T) {
	ok := checkStatus("", `["ns/conf/658248705"]`, "ns", "conf", 658248705)
	if !ok {
		t.Fail()
	}
}
