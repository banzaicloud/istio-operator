#!/usr/bin/env bash

dirname=$(dirname "$0")
projectdir=$PWD/$dirname/..
crdpath=$projectdir/deploy/charts/istio-operator/crds

ISTIO_VERSION=${1:-"1.10.3"}

for name in "$crdpath"/*.yaml; do
	sed "s/{{ISTIO_VERSION}}/${ISTIO_VERSION}/" "$projectdir/hack/crd-labeling-yq.yaml" | "$projectdir"/bin/yq w -i -s - "$name"
done
