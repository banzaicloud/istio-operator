#!/usr/bin/env bash
set +x
dirname=$(dirname "$0")
binpath=$PWD/$dirname/../bin
version="v0.18.6"
cmds="deepcopy-gen defaulter-gen lister-gen client-gen informer-gen"

for name in ${cmds}; do
    if [[ ! -f $binpath/$name ]]; then
        GOBIN=$binpath go get k8s.io/code-generator/cmd/"$name"@$version
    fi
done

if [[ ! -f $binpath/controller-gen1 ]]; then
    TMPDIR=$(mktemp -d -t ci-XXXXXXXXXX)
    if [ -d "$TMPDIR" ]; then
        BUILDDIR=$TMPDIR/cgen-build
        mkdir "$BUILDDIR"
        cp "$PWD"/"$dirname"/go.mod.cgen "$BUILDDIR"/go.mod
        pushd "$BUILDDIR" >/dev/null || exit
        GOBIN=$binpath GOMOD=$BUILDDIR/go.mod go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.1.9
        popd >/dev/null || exit
        rm -rf "$BUILDDIR"
    fi
fi

go mod vendor
