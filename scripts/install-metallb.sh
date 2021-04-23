#!/bin/bash

set -euo pipefail

readonly METALLB_VERSION=0.9.6

readonly LB_IP_OFFSET=${LB_IP_OFFSET:-255}
readonly CLEANUP=${CLEANUP:-false}

GLOBAL_METALLB_PREFIX=${GLOBAL_METALLB_PREFIX:-${LB_IP_OFFSET}}

function install_metallb {
    local node_addr_prefix
    node_addr_prefix=$(kubectl get nodes -o jsonpath="{.items[0].status.addresses[?(@.type=='InternalIP')].address}" | cut -d '.' -f 1,2)

    echo "Installing MetalLB: IP range: ${node_addr_prefix}.${GLOBAL_METALLB_PREFIX}.1-${node_addr_prefix}.${GLOBAL_METALLB_PREFIX}.250"

    kubectl apply -f https://raw.githubusercontent.com/google/metallb/v${METALLB_VERSION}/manifests/namespace.yaml
    kubectl apply -f https://raw.githubusercontent.com/google/metallb/v${METALLB_VERSION}/manifests/metallb.yaml

    if ! kubectl get secret -n metallb-system memberlist -o name &>/dev/null; then
        kubectl create secret generic -n metallb-system memberlist --from-literal=secretkey="$(openssl rand -base64 128)"
    fi

    eval "cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  namespace: metallb-system
  name: config
data:
  config: |
    address-pools:
    - name: default
      protocol: layer2
      addresses:
      - ${node_addr_prefix}.${GLOBAL_METALLB_PREFIX}.1-${node_addr_prefix}.${GLOBAL_METALLB_PREFIX}.250
EOF"

    echo "MetalLB installed"

    (( GLOBAL_METALLB_PREFIX-- )) || true
}

function remove_metallb {
    echo "Removing MetalLB"
    kubectl delete secret -n metallb-system memberlist
    kubectl delete -f https://raw.githubusercontent.com/google/metallb/v${METALLB_VERSION}/manifests/metallb.yaml
    kubectl delete -f https://raw.githubusercontent.com/google/metallb/v${METALLB_VERSION}/manifests/namespace.yaml
    echo "MetalLB removed"
}

if [[ ${CLEANUP} == true ]]; then
    remove_metallb
else
    install_metallb
fi
