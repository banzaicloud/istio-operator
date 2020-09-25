#!/usr/bin/env bash

set -euo pipefail

dirname=$(dirname "$0")
binpath=$PWD/$dirname/../bin
version="v0.18.6"
cmds="deepcopy-gen defaulter-gen lister-gen client-gen informer-gen"

for name in ${cmds}; do
    if [[ ! -f $binpath/$name ]]; then
        GOBIN=$binpath go get k8s.io/code-generator/cmd/"$name"@$version
    fi
done

cgen_version=0.2.9

target_name=controller-gen-${cgen_version}
link_path=${binpath}/controller-gen

[ -e "${link_path}" ] && rm -r "${link_path}"

if [ ! -e "${binpath}/${target_name}" ]; then
    GOBIN=$binpath go get sigs.k8s.io/controller-tools/cmd/controller-gen@v${cgen_version}
    mv "${binpath}/controller-gen" "${binpath}/${target_name}"
fi

ln -s "${target_name}" "${link_path}"

go mod vendor
