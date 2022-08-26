#!/usr/bin/env bash

set -euo pipefail

code_generator_version=v0.24.4
controller_gen_version=v0.9.2
istio_deps_version=1.15.0-beta.1
golang_protoc_grpc_version=v1.2.0
golang_protoc_gen_go_version=v1.28.0
istio_deps_replacement_version=v1.11.4-bzc.4
istio_tools_replacement_module_name=github.com/waynz0r/istio-tools
yq_version=v4.24.5

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
binpath=${script_dir}/../bin

function ensure-binary-version() {
    local bin_name=$1
    local bin_version=$2
    local download_location=$3
    local download_uri=$4
    local replacement_module_name=${5:-}

    local target_name=${bin_name}-proper-${bin_version}
    local link_path=${binpath}/${bin_name}

    if [ ! -L "${link_path}" ]; then
        rm -f "${link_path}"
    fi

    if [ ! -e "${binpath}/${target_name}" ]; then
        BUILD_DIR=$(mktemp -d)
        pushd "${BUILD_DIR}"
        go mod init foobar
        if [ ! -z "${replacement_module_name}" ]; then
            go mod edit -replace=${download_location}=${replacement_module_name}@${bin_version}
        fi
        cat << EOF > dummy.go
package main

import (
  _ "${download_location}${download_uri}"
)
EOF


        mkdir -p ${PWD}/bin

        GOBIN=${PWD}/bin go get "${download_location}${download_uri}@${bin_version}"
        GOBIN=${PWD}/bin go install "${download_location}${download_uri}"
        mkdir -p "${binpath}"
        mv "bin/${bin_name}" "${binpath}/${target_name}"
        popd
        rm -rf "${BUILD_DIR}"
        echo "${bin_name} ensured"
    fi

    ln -sf "${target_name}" "${link_path}"

}

# code generators
cmds="deepcopy-gen defaulter-gen lister-gen client-gen informer-gen"
for name in ${cmds}; do
    ensure-binary-version "${name}" ${code_generator_version} "k8s.io/code-generator" "/cmd/$name"
done

ensure-binary-version controller-gen ${controller_gen_version} "sigs.k8s.io/controller-tools" "/cmd/controller-gen"
ensure-binary-version cue-gen ${istio_deps_replacement_version} "istio.io/tools" "/cmd/cue-gen" "${istio_tools_replacement_module_name}"
ensure-binary-version protoc-gen-go-grpc ${golang_protoc_grpc_version} "google.golang.org/grpc" "/cmd/protoc-gen-go-grpc"
ensure-binary-version protoc-gen-go ${golang_protoc_gen_go_version} "google.golang.org/protobuf" "/cmd/protoc-gen-go"
ensure-binary-version protoc-gen-golang-deepcopy ${istio_deps_version} "istio.io/tools" "/cmd/protoc-gen-golang-deepcopy"
ensure-binary-version protoc-gen-golang-jsonshim ${istio_deps_version} "istio.io/tools" "/cmd/protoc-gen-golang-jsonshim"
ensure-binary-version protoc-gen-docs ${istio_deps_version} "istio.io/tools" "/cmd/protoc-gen-docs"
ensure-binary-version yq ${yq_version} "github.com/mikefarah/yq" "/v4"

go mod tidy
