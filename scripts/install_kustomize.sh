version=1.0.11 # versions starting from v2.0.0 have issues with outside paths: https://github.com/kubernetes-sigs/kustomize/issues/776
opsys=darwin  # or linux, or windows

# download the release
curl -O -L https://github.com/kubernetes-sigs/kustomize/releases/download/v${version}/kustomize_${version}_${opsys}_amd64

# move to /usr/local
mv kustomize_*_${opsys}_amd64 /usr/local/bin/kustomize
chmod u+x /usr/local/bin/kustomize