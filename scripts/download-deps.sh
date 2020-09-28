#!/usr/bin/env bash

set -euo pipefail

version="v0.18.6"
cgen_version=0.4.0
yq_version=3.4.0

dirname=$(dirname "$0")
binpath=$PWD/$dirname/../bin

# code generators
cmds="deepcopy-gen defaulter-gen lister-gen client-gen informer-gen"
for name in ${cmds}; do
    if [[ ! -f $binpath/$name ]]; then
        GOBIN=$binpath go get k8s.io/code-generator/cmd/"$name"@$version
    fi
done

# controller-gen
target_name=controller-gen-${cgen_version}
link_path=${binpath}/controller-gen

[ -e "${link_path}" ] && rm -r "${link_path}"

if [ ! -e "${binpath}/${target_name}" ]; then
    GOBIN=$binpath go get sigs.k8s.io/controller-tools/cmd/controller-gen@v${cgen_version}
    mv "${binpath}/controller-gen" "${binpath}/${target_name}"
fi

ln -s "${target_name}" "${link_path}"

# yq
if [[ ! -f $binpath/yq ]]; then
    GOBIN=$binpath go get github.com/mikefarah/yq/v3@${yq_version}
fi

go mod tidy
go mod vendor
