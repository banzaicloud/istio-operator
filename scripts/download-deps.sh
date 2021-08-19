#!/usr/bin/env bash

set -euo pipefail

code_generator_version=v0.22.0
controller_gen_version=v0.6.2
istio_deps_version=v1.11.0-bzc.1
istio_tools_replacement_module_name=github.com/waynz0r/istio-tools
gogo_protobuf_version=v1.3.2

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
        GOBIN=${PWD} go get "${download_location}${download_uri}@${bin_version}"
        mkdir -p "${binpath}"
        mv "${bin_name}" "${binpath}/${target_name}"
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
ensure-binary-version cue-gen ${istio_deps_version} "istio.io/tools" "/cmd/cue-gen" "${istio_tools_replacement_module_name}"
ensure-binary-version protoc-gen-deepcopy ${istio_deps_version} "istio.io/tools" "/cmd/protoc-gen-deepcopy" "${istio_tools_replacement_module_name}"
ensure-binary-version protoc-gen-jsonshim ${istio_deps_version} "istio.io/tools" "/cmd/protoc-gen-jsonshim" "${istio_tools_replacement_module_name}"
ensure-binary-version protoc-gen-docs ${istio_deps_version} "istio.io/tools" "/cmd/protoc-gen-docs" "${istio_tools_replacement_module_name}"
ensure-binary-version protoc-gen-gogofast ${gogo_protobuf_version} "github.com/gogo" "/protobuf/protoc-gen-gogofast"

go mod tidy
