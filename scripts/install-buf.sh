#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

target_name=buf-${version}
link_path=bin/buf

[ -e ${link_path} ] && rm -r ${link_path}

mkdir -p bin
ln -s "${target_name}" ${link_path}

if [ ! -e bin/"${target_name}" ]; then
    url="https://github.com/bufbuild/buf/releases/download/v${version}/buf-$(uname -s)-$(uname -m)"
    curl -s -L "${url}" -o bin/"${target_name}"
    chmod u+x bin/"${target_name}"
fi
