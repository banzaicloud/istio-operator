#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

target_name=kustomize-${version}
link_path=bin/kustomize

[ -e ${link_path} ] && rm -r ${link_path}

mkdir -p bin
ln -s "${target_name}" ${link_path}

if [ ! -e bin/"${target_name}" ]; then
    os=$(go env GOOS)
    arch=$(go env GOARCH)

    # Temporary fix for Apple M1 until kustomize is released for darwin-arm64 arch
    if [ "$os" == "darwin" ] && [ "$arch" == "arm64" ]; then
        arch="amd64"
    fi

    url="https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v${version}/kustomize_v${version}_${os}_${arch}.tar.gz"
    curl -L "${url}" | tar -xz -C /tmp/
    mv "/tmp/kustomize" bin/"${target_name}"
    chmod u+x bin/"${target_name}"
fi
