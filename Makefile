# Image URL to use all building/pushing image targets
TAG ?= $(shell git describe --tags --abbrev=0 --match '[0-9].*[0-9].*[0-9]' 2>/dev/null )
IMAGE_REPOSITORY ?= banzaicloud/istio-operator
IMG ?= ${IMAGE_REPOSITORY}:$(TAG)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS = "crd:trivialVersions=true,maxDescLen=0,preserveUnknownFields=false,allowDangerousTypes=true"

TEST_RACE_DETECTOR ?= 0

RELEASE_TYPE ?= p
RELEASE_MSG ?= "operator release"

REL_TAG = $(shell ./scripts/increment_version.sh -${RELEASE_TYPE} ${TAG})

GOLANGCI_VERSION = 1.31.0
LICENSEI_VERSION = 0.1.0
KUBEBUILDER_VERSION = 2.3.1
KUSTOMIZE_VERSION = 2.0.3
ISTIO_VERSION = 1.9.3

KUSTOMIZE_BASE = config/overlays/specific-manager-version

ifeq (${TEST_RACE_DETECTOR}, 1)
    TEST_GOARGS += -race
    TEST_CGO_ENABLED = 1
endif
TEST_CGO_ENABLED ?= 0

all: test manager

.PHONY: check
check: test lint shellcheck-makefile shellcheck check-diff ## Run tests and linters

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ./bin v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
	@bin/golangci-lint run -v

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-check
license-check: bin/licensei ## Run license check
	bin/licensei check
	./scripts/check-header.sh

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	bin/licensei cache

# Run tests
.PHONY: test
test: install-kubebuilder generate fmt vet manifests
	env CGO_ENABLED=${TEST_CGO_ENABLED} \
	    KUBEBUILDER_ASSETS="$${PWD}/bin/kubebuilder/bin" \
	    go test ${TEST_GOARGS} ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet build

# Build manager binary
build:
	go build -o bin/manager -ldflags="-X github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1.SupportedIstioVersion=${ISTIO_VERSION} -X github.com/banzaicloud/istio-operator/pkg/apis/istio/v1beta1.Version=${TAG}" github.com/banzaicloud/istio-operator/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install kustomize
install-kustomize:
	scripts/install_kustomize.sh ${KUSTOMIZE_VERSION}

# Install kubebuilder
install-kubebuilder:
	scripts/install_kubebuilder.sh ${KUBEBUILDER_VERSION}

.PHONY: vendor
vendor:
	go mod vendor

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/base/crds

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install-kustomize
	bin/kustomize build config | kubectl apply -f -
	./scripts/image_patch.sh "${KUSTOMIZE_BASE}/manager_image_patch.yaml" ${IMG}
	bin/kustomize build $(KUSTOMIZE_BASE) | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: download-deps
	bin/controller-gen $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:rbac:artifacts:config=config/base/rbac output:crd:artifacts:config=config/base/crds
	rm deploy/charts/istio-operator/crds/* && cp config/base/crds/* deploy/charts/istio-operator/crds/
	scripts/label-crds.sh $(ISTIO_VERSION)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

download-deps:
	./scripts/download-deps.sh

# Generate code
generate: download-deps
	go generate ./pkg/... ./cmd/...
	bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
	./hack/update-codegen.sh
	go run static/generate.go

# Verify codegen
verify-codegen: download-deps
	./hack/verify-codegen.sh

# Build the docker image
docker-build: vendor
	docker build -f Dockerfile.dev . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

.PHONY: e2e-test-dependencies
e2e-test-dependencies:
	./scripts/e2e-test/download-deps.sh

.PHONY: e2e-test-env
e2e-test-env: export PATH:=./bin:${PATH}
e2e-test-env: e2e-test-dependencies
	kind create cluster

	# TODO get these from the MetalLB resource yaml
	docker pull metallb/controller:v0.9.6
	docker pull metallb/speaker:v0.9.6
	kind load docker-image metallb/controller:v0.9.6
	kind load docker-image metallb/speaker:v0.9.6
	./scripts/install-metallb.sh

	docker pull kennethreitz/httpbin
	kind load docker-image kennethreitz/httpbin

	# TODO collect these from the operator (run the related functions and collect the referenced images)
	docker pull gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
	docker pull docker.io/istio/pilot:${ISTIO_VERSION}
	docker pull docker.io/istio/proxyv2:${ISTIO_VERSION}
	kind load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
	kind load docker-image docker.io/istio/pilot:${ISTIO_VERSION}
	kind load docker-image docker.io/istio/proxyv2:${ISTIO_VERSION}

.PHONY: e2e-test-install-istio-operator
e2e-test-install-istio-operator: export PATH:=./bin:${PATH}
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

	# TODO maybe wait until all pods are up and running?
	#  `helm --wait` only waits for the pods installed by helm. Usually, when the istio-operator is ready,
	#  a couple of pods in kube-system are still just starting up. It works out fine now, probably because
	#  there is a wait in TestMain for the cluster to be reachable. Waiting here for all pods might result
	#  in a lower load on the cluster when the actual tests start, so it might remove some flakiness.

.PHONY: e2e-test
e2e-test: export PATH:=./bin:${PATH}
e2e-test: e2e-test-install-istio-operator
	env CGO_ENABLED=${TEST_CGO_ENABLED} \
	    go test --failfast --timeout 10m -v ${TEST_GOARGS} ./test/e2e/... -coverprofile cover.out

    # TODO collect used docker images and compare with known list. This list can be used to preload the images into kind
    # TODO  `kind export logs` and look for "ImageCreate" in containerd.log

check_release:
	@echo "A new tag (${REL_TAG}) will be pushed to Github, and a new Docker image will be released. Are you sure? [y/N] " && read -r ans && [ "$${ans:-N}" = y ]

.PHONY: check-diff
check-diff:
	@git --no-pager diff --name-only --exit-code -- 'static/*' || (echo "Please commit any changes to the generated code" && false)

release: check_release
	git tag -a ${REL_TAG} -m ${RELEASE_MSG}
	git push origin ${REL_TAG}

.PHONY: shellcheck-makefile
shellcheck-makefile: bin/shellcheck ## Check each makefile recipe using shellcheck
	@grep -h -E '^[a-zA-Z_-]+:' $(MAKEFILE_LIST) | cut -d: -f1 | while IFS= read -r target; do \
		echo "Checking make target: $$target"; \
		make -s "$$target" SHELL=scripts/shellcheck-makefile.sh || exit 1; \
	done

.PHONY: shellcheck
shellcheck: bin/shellcheck ## Check shell scripts
	bin/shellcheck scripts/*.sh hack/*.sh docs/federation/gateway/cluster-add/*.sh docs/federation/flat/*.sh

bin/shellcheck:
	scripts/install_shellcheck.sh
