#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

# install limiter
make undeploy

kubectl delete -f https://raw.githubusercontent.com/istio/istio/master/samples/addons/prometheus.yaml

kubectl delete -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
kubectl delete -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/sample-client/fortio-deploy.yaml

istioctl x uninstall --purge -y

kubectl delete ns istio-system

