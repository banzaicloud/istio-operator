#!/usr/bin/env bash

set -euo pipefail

version=0.7.1
arch=$(uname -m)
opsys=$(uname -s | awk '{print tolower($0)}')

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
BIN_DIR=${SCRIPT_DIR}/../bin
mkdir -p "${BIN_DIR}"

# download the release
curl -L -O "https://github.com/koalaman/shellcheck/releases/download/v${version}/shellcheck-v${version}.${opsys}.${arch}.tar.xz"

# extract the archive
tar -Jxvf "shellcheck-v${version}.${opsys}.${arch}.tar.xz" shellcheck-v${version}/shellcheck
mv shellcheck-v${version}/shellcheck "${BIN_DIR}"
rmdir shellcheck-v${version}

# delete tar file
rm "shellcheck-v${version}.${opsys}.${arch}.tar.xz"
