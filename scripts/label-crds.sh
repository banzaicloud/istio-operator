#!/usr/bin/env bash

dirname=$(dirname "$0")
projectdir=$PWD/$dirname/..
crdpath=$projectdir/config/crd/bases

ISTIO_VERSION=${1:-"1.13.3"}

for name in "$crdpath"/*.yaml; do
	sed "$ d" $name > $name.changed
	mv $name.changed $name

    "$projectdir"/bin/yq ".metadata.labels.\"resource.alpha.banzaicloud.io/revision\" = \"$ISTIO_VERSION\"" -i "$name"
done
