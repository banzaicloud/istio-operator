#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

target_dir_name=kubebuilder-${version}
link_path=bin/kubebuilder

[ -e ${link_path} ] && rm -r ${link_path}

mkdir -p bin
ln -s "${target_dir_name}" ${link_path}

if [ ! -e bin/"${target_dir_name}" ]; then
    os=$(go env GOOS)
    arch=$(go env GOARCH)

    # download kubebuilder and extract it to tmp
    curl -L "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${version}/kubebuilder_${version}_${os}_${arch}.tar.gz" | tar -xz -C /tmp/
    mv "/tmp/kubebuilder_${version}_${os}_${arch}" bin/"${target_dir_name}"
fi
