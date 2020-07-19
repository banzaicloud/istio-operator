#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

target_name=kustomize-${version}
link_path=bin/kustomize

if [ -e ${link_path} ] && [ ! -L ${link_path} ]; then
    echo "Please move ${link_path} out of the way"
    exit 1
fi

mkdir -p bin
rm -f ${link_path}
ln -s "${target_name}" ${link_path}

if [ ! -e bin/"${target_name}" ]; then
    os=$(go env GOOS)
    arch=$(go env GOARCH)

    url="https://github.com/kubernetes-sigs/kustomize/releases/download/v${version}/kustomize_${version}_${os}_${arch}"
    curl -L "${url}" -o bin/"${target_name}"
    chmod u+x bin/"${target_name}"
fi
