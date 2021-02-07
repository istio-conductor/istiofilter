#!/usr/bin/env bash
set -e
BASEDIR=$(realpath $(dirname "$0"))
PROJECT_DIR=$(realpath "${BASEDIR}/../")
export FILTER_VERSION=${FILTER_VERSION:-latest}
export NAMESPACE=istio-system

# change current dir to demo/
cd ${BASEDIR}
kubectl apply -f ${PROJECT_DIR}/kubernetes/customresourcedefinitions.gen.yaml

# check if certs exist
EXIST=1
kubectl get secret istiofilter-certs -n ${NAMESPACE} -o jsonpath='{.data}' > ./cert.json|| EXIST=0

if [ $EXIST == 0 ]; then
  echo 'generating istiofilter certs:'
  openssl req -x509 -sha256 -nodes -newkey rsa:2048  -keyout key.pem -out cert.pem -days 36500 \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=istio-conductor.org/OU=Org/CN=istiofilter.istio-system.svc" \
  -addext "subjectAltName = DNS:istiofilter.istio-system.svc"
else
  cat cert.json|jq -r '."cert.pem"'|base64 -d > cert.pem
  cat ./cert.json|jq -r '."key.pem"'|base64 -d > key.pem
fi

export KEY=$(cat key.pem|base64)
export CERT=$(cat cert.pem|base64)
envsubst < ${PROJECT_DIR}/test/e2e/common/istiofilter.yaml > istiofilter.yaml
kubectl apply -f istiofilter.yaml
