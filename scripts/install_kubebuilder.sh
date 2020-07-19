#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

target_dir_name=kubebuilder-${version}
link_path=bin/kubebuilder

if [ -e ${link_path} ] && [ ! -L ${link_path} ]; then
    echo "Please move ${link_path} out of the way"
    exit 1
fi

mkdir -p bin
rm -f ${link_path}
ln -s "${target_dir_name}" ${link_path}

if [ ! -e bin/"${target_dir_name}" ]; then
    os=$(go env GOOS)
    arch=$(go env GOARCH)

    # download kubebuilder and extract it to tmp
    curl -L "https://go.kubebuilder.io/dl/${version}/${os}/${arch}" | tar -xz -C /tmp/

    # extract the archive
    mv "/tmp/kubebuilder_${version}_${os}_${arch}" bin/"${target_dir_name}"
fi
