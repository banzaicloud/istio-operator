#!/bin/bash

set -euo pipefail

readonly KUBECTL_VERSION=1.20.4
readonly HELM_VERSION=3.2.3
readonly KIND_VERSION=0.10.0

readonly script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly repo_root=${script_dir}/../../../
readonly bin_path=${repo_root}/bin

mkdir -p "${bin_path}"

readonly kernel=$(uname -s |tr '[:upper:]' '[:lower:]')

function download_naked() {
    local name=$1
    local version=$2
    # shellcheck disable=SC2034
    local unused_param=$3
    local url=$4

    local target="${bin_path}/${name}-${version}"

    curl -L "${url}" -o "${target}"
    chmod 755 "${target}"
}

function download_tar_gz() {
    local name=$1
    local version=$2
    local file_in_archive=$3
    local url=$4

    local target="${bin_path}/${name}-${version}"

    local tmpdir
    tmpdir=$(mktemp -d)

    curl -L "${url}" | tar xz -C "${tmpdir}" "${file_in_archive}"
    mv "${tmpdir}/${file_in_archive}" "${target}"
    chmod 755 "${target}"

    rm -r "${tmpdir}"
}

function ensure() {
    local name=$1
    local version=$2
    local download_method=$3
    local download_param=$4
    local url=$5

    local target="${bin_path}/${name}-${version}"

    if [ ! -x "${target}" ]; then
        "${download_method}" "${name}" "${version}" "${download_param}" "${url}"
    fi
    ln -sf "${name}-${version}" "${bin_path}/${name}"
}

ensure kubectl ${KUBECTL_VERSION} download_naked "" "https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/${kernel}/amd64/kubectl"
ensure helm ${HELM_VERSION} download_tar_gz "${kernel}-amd64/helm" "https://get.helm.sh/helm-v${HELM_VERSION}-${kernel}-amd64.tar.gz"
ensure kind ${KIND_VERSION} download_naked "" "https://github.com/kubernetes-sigs/kind/releases/download/v${KIND_VERSION}/kind-${kernel}-amd64"
