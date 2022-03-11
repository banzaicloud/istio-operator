#!/usr/bin/env bash

dirname=$(dirname "$0")
projectdir=$PWD/$dirname/..
crdpath=$projectdir/config/crd/bases

ISTIO_VERSION=${1:-"1.12.5"}

for name in "$crdpath"/*.yaml; do
	sed "$ d" $name > $name.changed
	mv $name.changed $name
	sed "s/{{ISTIO_VERSION}}/${ISTIO_VERSION}/" "$projectdir/hack/crd-labeling-yq.yaml" | "$projectdir"/bin/yq w -d'*' -i -s - "$name"
done
