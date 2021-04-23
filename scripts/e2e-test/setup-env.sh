#!/bin/bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <istio-version>"; exit 1; }

readonly istio_version=$1

readonly script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly scripts_dir=${script_dir}/..

kind create cluster

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
docker pull docker.io/istio/pilot:"${istio_version}"
docker pull docker.io/istio/proxyv2:"${istio_version}"
kind load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
kind load docker-image docker.io/istio/pilot:"${istio_version}"
kind load docker-image docker.io/istio/proxyv2:"${istio_version}"
