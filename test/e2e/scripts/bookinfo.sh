#!/usr/bin/env bash
BASEDIR=$(realpath $(dirname "$0"))

kubectl apply -f ${BASEDIR}/../bookinfo/bookinfo.yaml
kubectl apply -f ${BASEDIR}/../bookinfo/destination-rule-all.yaml
kubectl apply -f ${BASEDIR}/../bookinfo/virtual-service-all-v1.yaml