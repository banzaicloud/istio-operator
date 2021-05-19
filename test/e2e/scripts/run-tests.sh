#!/bin/bash

set -euo pipefail

readonly script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly repo_root=${script_dir}/../../..

export PATH=${repo_root}/bin:${PATH}

readonly log_dir=${E2E_LOG_DIR:-${PWD}/logs}
readonly fail_fast=${E2E_TEST_FAIL_FAST:-0}
readonly ginkgo_args=${E2E_TEST_GINKGO_ARGS:-}

mkdir -p "${log_dir}"
env E2E_TEST_FAIL_FAST="${fail_fast}" \
    E2E_LOG_DIR="${log_dir}" \
    E2E_TEST_DUMP_SCRIPT="${repo_root}"/scripts/dump-cluster-state-and-logs.sh \
    ginkgo "${ginkgo_args}" --randomizeSuites --randomizeAllSpecs --timeout 10m -v ./test/e2e/... \
        | tee "${log_dir}"/e2e-test.log

# TODO collect used docker images and compare with known list. This list can be used to preload the images into kind
# TODO  `kind export logs` and look for "ImageCreate" in containerd.log
