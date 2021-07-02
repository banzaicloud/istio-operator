#!/usr/bin/env bash

set -euo pipefail

[ -z "${1:-}" ] && { echo "Usage: $0 <version>"; exit 1; }

version=$1

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
build_dir=${script_dir}/../build

${script_dir}/remove-istio-dependencies.sh ${build_dir}

pushd ${build_dir}

echo "clone istio api repository"
git clone -q -b ${version} https://github.com/istio/api

echo "copy dependencies"
cp -a api/mesh api/networking .
cp -a api/operator/fixup_structs .

for i in `ls -1 api/common-protos`; do cp -a api/common-protos/$i $i; done

find mesh networking -type f -not -name '*.proto' -exec rm {} \;


rm -rf api

ln -s ../api api

popd
