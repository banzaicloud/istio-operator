#!/usr/bin/env bash

version=2.0.3
opsys=$(uname -s | awk '{print tolower($0)}')

# download the release
curl -O -L "https://github.com/kubernetes-sigs/kustomize/releases/download/v${version}/kustomize_${version}_${opsys}_amd64"

# move to bin
mkdir -p bin
mv "kustomize_*_${opsys}_amd64" bin/kustomize
chmod u+x bin/kustomize
