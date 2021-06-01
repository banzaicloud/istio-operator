# When running the end-to-end tests, an istio-operator docker image is built from the local
# source, then loaded into the kind cluster as `banzaicloud/istio-operator:${E2E_TEST_TAG}`.
# The reason for using this variable instead of `TAG` is to make sure the pod isn't showing
# a released version tag (the `TAG` variable defaults to the latest released version).
# Without this variable, the installed istio-operator pod would contain something like
# `image: banzaicloud/istio-operator:0.9.3`, but it would _not_ be running the 0.9.3 version
# of istio-operator. On the other hand, if for some reason the image is not loaded into the
# cluster before installing the istio-operator, the `banzaicloud/istio-operator:0.9.3` image
# would be pulled down and started, which wouldn't be the expected image.
# To avoid this and similar issues, the tag is forced to be some other value: `e2e-test` by
# default, or the current branch name on the CI.
# Currently, it's not possible to run the tests with a different istio-operator version.
E2E_TEST_TAG ?= e2e-test

E2E_LOG_DIR ?= ${PWD}/logs

E2E_TEST_FAIL_FAST ?= 0

ifeq (${E2E_TEST_FAIL_FAST}, 1)
	E2E_TEST_GINKGO_ARGS += --failFast
endif

.PHONY: e2e-test-dependencies
e2e-test-dependencies:
	./test/e2e/scripts/download-deps.sh

.PHONY: e2e-test-env
e2e-test-env: e2e-test-dependencies
	# There's an issue (https://github.com/banzaicloud/istio-operator/issues/643) with resource cleanup on
	# k8s 1.20, so running the tests on 1.19.7 for now
	./test/e2e/scripts/setup-env.sh 1.19.7 ${ISTIO_VERSION}

.PHONY: e2e-test-install-istio-operator
e2e-test-install-istio-operator: export PATH:=./bin:${PATH}
e2e-test-install-istio-operator: TAG=${E2E_TEST_TAG}
e2e-test-install-istio-operator: docker-build
	# TODO build with TEST_RACE_DETECTOR=1 in docker-build
	kind load docker-image ${IMG}
	helm install --wait \
		--set operator.image.repository=${IMAGE_REPOSITORY} \
		--set operator.image.tag=${TAG} \
		--create-namespace \
		--namespace istio-system \
		istio-operator-e2e-test \
		deploy/charts/istio-operator/

	# Wait for all pods to be ready. `helm --wait` only waits for the pods installed by helm. Usually,
	# when the istio-operator is ready, a couple of pods in kube-system are still just starting up. It
	# works out fine now, probably because there is a wait in TestMain for the cluster to be reachable.
	# Waiting here for all pods might result in a lower load on the cluster when the actual tests
	# start, so it might remove some flakiness.
	kubectl wait pod --all-namespaces --all --for=condition=ready --timeout=60s

.PHONY: e2e-test-run
e2e-test-run:
	mkdir -p ${E2E_LOG_DIR}
	env E2E_TEST_FAIL_FAST=${E2E_TEST_FAIL_FAST} E2E_LOG_DIR=${E2E_LOG_DIR} \
		E2E_TEST_DUMP_SCRIPT=${PWD}/scripts/dump-cluster-state-and-logs.sh \
		bin/ginkgo ${E2E_TEST_GINKGO_ARGS} --randomizeSuites --randomizeAllSpecs --timeout 10m -v ./test/e2e/... \
			| tee ${E2E_LOG_DIR}/e2e-test.log

    # TODO collect used docker images and compare with known list. This list can be used to preload the images into kind
    # TODO  `kind export logs` and look for "ImageCreate" in containerd.log

.PHONY: e2e-test
e2e-test: download-deps e2e-test-install-istio-operator e2e-test-run
