#!/bin/bash

set -euo pipefail

if [[ $# != 2 || $1 != "-c" ]]; then
    echo "Usage: $0 -c <commands>"
    exit 1
fi

echo "$2" | shellcheck -s sh -
