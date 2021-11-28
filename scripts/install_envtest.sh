#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

target_dir_name=envtest-${version}
link_path=bin/envtest

[ -L ${link_path} ] && rm -r ${link_path}

mkdir -p bin
ln -s "${target_dir_name}" ${link_path}

if [ ! -e bin/"${target_dir_name}" ]; then
    curl -sSL "https://go.kubebuilder.io/test-tools/${version}/$(go env GOOS)/$(go env GOARCH)" | tar -xz -C /tmp/
    mv "/tmp/kubebuilder" bin/"${target_dir_name}"
fi
