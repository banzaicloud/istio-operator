#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <directory>"; exit 1; }

build_dir=$1

pushd ${build_dir}

echo "cleanup"
rm -rf api common-protos github.com gogoproto google istio.io k8s.io mesh networking ../mesh ../networking

popd
