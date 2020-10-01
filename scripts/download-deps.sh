#!/usr/bin/env bash

set -euo pipefail

code_generator_version=0.0.0-20180823001027-3dcf91f64f63
controller_gen_version=0.1.9

dirname=$(dirname "$0")
binpath=$PWD/$dirname/../bin

function ensure-binary-version() {
    local bin_name=$1
    local bin_version=$2
    local download_location=$3

    local target_name=${bin_name}-${bin_version}
    local link_path=${binpath}/${bin_name}

    [ -e "${link_path}" ] && rm "${link_path}"

    if [ ! -e "${binpath}/${target_name}" ]; then
        GOBIN=$binpath go get "${download_location}"
        mv "${binpath}/${bin_name}" "${binpath}/${target_name}"
    fi

    ln -s "${target_name}" "${link_path}"
}

# code generators
cmds="deepcopy-gen defaulter-gen lister-gen client-gen informer-gen"
for name in ${cmds}; do
    ensure-binary-version "${name}" ${code_generator_version} "k8s.io/code-generator/cmd/$name@v${code_generator_version}"
done

ensure-binary-version controller-gen ${controller_gen_version} "sigs.k8s.io/controller-tools/cmd/controller-gen@v${controller_gen_version}"

go mod tidy
go mod vendor
