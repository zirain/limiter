#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


# For all commands, the working directory is the parent directory(repo root).
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "${REPO_ROOT}"
source "${REPO_ROOT}"/hack/util.sh

if ! [ -x "$(command -v kustomize)" ]; then
    util::install_tools sigs.k8s.io/kustomize/kustomize/v4 v4.5.5
fi

if ! [ -x "$(command -v setup-envtest)" ]; then
    util::install_tools sigs.k8s.io/controller-runtime/tools/setup-envtest latest
fi

if ! [ -x "$(command -v ginkgo)" ]; then
    util::install_tools github.com/onsi/ginkgo/v2/ginkgo v2.0.0
fi

if ! [ -x "$(command -v controller-gen)" ]; then
    util::install_tools sigs.k8s.io/controller-tools/cmd/controller-gen v0.8.0
fi

if ! [ -x "$(command -v deepcopy-gen)" ]; then
    util::install_tools k8s.io/code-generator/cmd/deepcopy-gen latest
fi

if ! [ -x "$(command -v register-gen)" ]; then
    util::install_tools k8s.io/code-generator/cmd/register-gen latest
fi

if ! [ -x "$(command -v client-gen)" ]; then
    util::install_tools k8s.io/code-generator/cmd/client-gen latest
fi

if ! [ -x "$(command -v lister-gen)" ]; then
    util::install_tools k8s.io/code-generator/cmd/lister-gen latest
fi

if ! [ -x "$(command -v informer-gen)" ]; then
    util::install_tools k8s.io/code-generator/cmd/informer-gen latest
fi