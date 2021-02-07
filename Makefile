SHELL := /bin/bash

ISTIO_VERSION?=1.8.2

CMD_ENTRY=./cmd/istiofilter
TEST_PATH?=./test/e2e/
OUT=./out
BINARY=$(OUT)/istiofilter
LINUX_BINARY=$(OUT)/linux/istiofilter
DOCKER_BUILD_DIR=$(OUT)/docker
DOCKER_TAG?=istioconductor/istiofilter:latest

prepare:
	mkdir -p $(DOCKER_BUILD_DIR)

linux_binary: prepare
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(LINUX_BINARY) $(CMD_ENTRY)

docker: linux_binary
	cp $(LINUX_BINARY) $(DOCKER_BUILD_DIR)
	cp ./Dockerfile $(DOCKER_BUILD_DIR)
	cd $(DOCKER_BUILD_DIR) && docker build -f ./Dockerfile -t $(DOCKER_TAG) .

local_binary:
	go build -o $(BINARY) $(CMD_ENTRY)

e2e_local_prepare: docker
	sh $(TEST_PATH)scripts/istio.sh -y -f $(TEST_PATH)/common/istio-config.yaml
	sh $(TEST_PATH)scripts/istiofilter.sh

.PHONY: prepare docker local_binary linux_binary e2e_local_prepare
