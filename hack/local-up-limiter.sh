#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

istioctl version

istioctl install -f "$REPO_ROOT/hack/iop/demo.yaml" -y



kubectl label namespace default istio-injection=enabled --overwrite

kubectl apply -f "$REPO_ROOT/hack/prometheus/prometheus.yaml"

kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/httpbin.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/master/samples/httpbin/sample-client/fortio-deploy.yaml

# install limiter
make deploy