#!/usr/bin/env bash

set -euo pipefail

readonly dump_dir=${1:-}

[ -z "${dump_dir}" ] && { echo "Usage: $0 <dump-dir>"; exit 1; }

mkdir -p "${dump_dir}"

function dump-cluster-state {
    echo "dumping cluster state" > "${dump_dir}"/cluster-state
}

function dump-logs {
    echo "dumping logs" > "${dump_dir}"/logs
}

dump-cluster-state
dump-logs
