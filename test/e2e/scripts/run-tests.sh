#!/bin/bash

set -euo pipefail

readonly script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
readonly repo_root=${script_dir}/../../..

export PATH=${repo_root}/bin:${PATH}

readonly log_dir=${E2E_LOG_DIR:-${PWD}/logs}
readonly fail_fast=${E2E_TEST_FAIL_FAST:-0}
readonly ginkgo_args=${E2E_TEST_GINKGO_ARGS:-}

# Mostly for testing the "in docker way" on Linux
readonly force_run_in_docker=${E2E_TEST_FORCE_RUN_IN_DOCKER:-0}

mkdir -p "${log_dir}"

function run-tests-directly() {
    echo "Running tests directly"
    env E2E_TEST_FAIL_FAST="${fail_fast}" \
        E2E_LOG_DIR="${log_dir}" \
        E2E_TEST_DUMP_SCRIPT="${repo_root}"/scripts/dump-cluster-state-and-logs.sh \
        ginkgo "${ginkgo_args}" --randomizeSuites --randomizeAllSpecs --timeout 10m -v ./test/e2e/... \
            | tee "${log_dir}"/e2e-test.log
}

function run-tests-in-docker() {
    # this doesn't work at the moment. Two kind of connections are necessary for the tests:
    # 1. kind needs to be accessible. This works from the host machine, and it's probably accessible from a docker
    #    container too, but I'm not sure how to make it work.
    # 2. the IP range MetalLB is using needs to be accessible. This only works inside docker on the "kind" network.
    #    Note: there are "solutions" to make it accessible from the host machine, but what I've seen are ugly hacks
    #    which are not scriptable.
    echo "Running tests in docker container"

    local tmpdir
    tmpdir=$(mktemp -d)

    # KUBECONFIG can contain multiple paths which makes it hard to share the referenced files with the docker container.
    # To avoid this problem, the kubeconfig is exported into a dir which is then shared with the docker container.
    kind export kubeconfig --kubeconfig "${tmpdir}"/config.yaml

    docker build -f "${repo_root}"/test/e2e/Dockerfile "${repo_root}" -t istio-operator-e2e-test-runner
    docker run -t \
        --network kind \
        -v "${log_dir}":/logs \
        -v "${tmpdir}":/kind-config \
        -e E2E_LOG_DIR="${log_dir}" \
        -e E2E_TEST_FAIL_FAST="${fail_fast}" \
        -e E2E_TEST_GINKGO_ARGS="${ginkgo_args}" \
        -e KUBECONFIG=/kind-config/config.yaml \
        istio-operator-e2e-test-runner

    rm -r "${tmpdir}"
}

if [ "$(uname)" != "Linux" ] || [ "${force_run_in_docker}" -eq 1 ]; then
        run-tests-in-docker
else
        run-tests-directly
fi

# TODO collect used docker images and compare with known list. This list can be used to preload the images into kind
# TODO  `kind export logs` and look for "ImageCreate" in containerd.log
