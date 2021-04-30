#!/usr/bin/env bash

set -euo pipefail

readonly dump_dir=${1:-}

readonly script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly bin_path=${script_dir}/../bin
export PATH=${bin_path}:$PATH

[ -z "${dump_dir}" ] && {
    echo "Usage: $0 <dump-dir>"
    exit 1
}

mkdir -p "${dump_dir}"

function dump-cluster-state() {
    local dir="${dump_dir}"/resources
    mkdir -p "${dir}"

    # This list should probably match the list in test/e2e/e2e_helpers_test.go:listAllResources()
    local -a kinds=(
        crd service pod deployment horizontalpodautoscaler clusterrole clusterrolebinding
        validatingwebhookconfiguration mutatingwebhookconfiguration destinationrule
        virtualservice peerauthentication gateway envoyfilter istio meshgateway
    )
    for kind in "${kinds[@]}"; do
        kubectl get "$kind" -A -o yaml > "$dir/$kind".yaml
    done
}

function dump-logs() {
    local namespaces
    namespaces=$(kubectl get namespace -o name)
    for ns in ${namespaces[*]}; do
        ns=${ns##namespace/}
        for pod in $(kubectl get -n "$ns" pod -o name); do
            pod=${pod##pod/}
            for container in $(kubectl get -n "$ns" pod "$pod" -o template='{{range .spec.containers}}{{.name}}{{"\n"}}{{end}}'); do
                local dir="${dump_dir}/logs/$ns/$pod"
                mkdir -p "$dir"
                kubectl logs -n "$ns" "$pod" -c "$container" >"$dir/$container.log"
            done
        done
    done
}

dump-cluster-state
dump-logs
