package controller

import "testing"

func TestCheckStatus(t *testing.T) {
	ok := checkStatus("", `["go-grpc-test-core/go-grpc-test-core-rpc/658248705"]`, "go-grpc-test-core", "go-grpc-test-core-rpc", 658248705)
	if !ok {
		t.Fail()
	}
}
