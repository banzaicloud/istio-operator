#!/usr/bin/env bash

set -euo pipefail

binpath=${PWD}/$(dirname "$0")/../bin
version="v0.0.0-20180823001027-3dcf91f64f63"
cmds="deepcopy-gen defaulter-gen lister-gen client-gen informer-gen"

for name in ${cmds}; do
    if [[ ! -f $binpath/$name ]]; then
        GOBIN=$binpath go get k8s.io/code-generator/cmd/"$name"@$version
    fi
done

if [[ ! -f $binpath/controller-gen ]]; then
    GOBIN=$binpath go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.1.9
fi

go mod vendor
