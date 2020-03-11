#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
BIN_DIR=${SCRIPT_DIR}/../bin
mkdir -p "${BIN_DIR}"

if [[ $# != 2 || $1 != "-c" ]]; then
    echo "Usage: $0 -c <commands>"
    exit 1
fi

echo "$2" | "${BIN_DIR}/shellcheck" -s sh -
