#!/bin/bash

set -euo pipefail

readonly kubernetes_version=${1:-}
readonly istio_version=${2:-}

if [ -z "${kubernetes_version}" ] || [ -z "${istio_version}" ]; then
    echo "Usage: $0 <kubernetes-version> <istio-version>"
    echo "Note: <kubernetes-version> must be a version for which there is a KinD node image, e.g. 1.19.7"
    echo "      Look for supported versions at https://hub.docker.com/r/kindest/node/tags"
    exit 1
fi

readonly script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly repo_root=${script_dir}/../../../
readonly scripts_dir=${repo_root}/scripts

export PATH=${repo_root}/bin:${PATH}

kind create cluster --image "kindest/node:v${kubernetes_version}"

# TODO get these from the MetalLB resource yaml
docker pull metallb/controller:v0.9.6
docker pull metallb/speaker:v0.9.6
kind load docker-image metallb/controller:v0.9.6
kind load docker-image metallb/speaker:v0.9.6
"${scripts_dir}"/install-metallb.sh

docker pull kennethreitz/httpbin
kind load docker-image kennethreitz/httpbin

# TODO collect these from the operator (run the related functions and collect the referenced images)
docker pull gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
docker pull gcr.io/istio-release/pilot:"${istio_version}"
docker pull gcr.io/istio-release/proxyv2:"${istio_version}"
kind load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
kind load docker-image gcr.io/istio-release/pilot:"${istio_version}"
kind load docker-image gcr.io/istio-release/proxyv2:"${istio_version}"
