# Image URL to use all building/pushing image targets
TAG ?= $(shell git describe --tags --abbrev=0 --match 'v[0-9].*[0-9].*[0-9]' 2>/dev/null )
IMAGE_REPOSITORY ?= ghcr.io/banzaicloud/istio-operator
IMG ?= ${IMAGE_REPOSITORY}:$(TAG)

CHART_VERSION ?= $(shell sed -nr '/version:/ s/.*version: ([^"]+).*/\1/p' deploy/charts/istio-operator/Chart.yaml)

RELEASE_TYPE ?= p
RELEASE_MSG ?= "istio operator release"
API_RELEASE_MSG ?= "istio operator api release"
EMBEDDED_CHART_RELEASE_MSG ?= "istio operator embedded chart release"
HELM_CHART_RELEASE_MSG ?= "istio operator helm chart release"

REL_TAG = $(shell ./scripts/increment_version.sh -${RELEASE_TYPE} ${TAG})
API_REL_TAG ?= api/${REL_TAG}
EMBEDDED_CHART_REL_TAG ?= deploy/charts/v${CHART_VERSION}
HELM_CHART_REL_TAG ?= chart/istio-operator/${CHART_VERSION}

GOLANGCI_VERSION = 1.45.2
LICENSEI_VERSION = 0.7.0
ENVTEST_K8S_VERSION = 1.24.2
KUSTOMIZE_VERSION = 4.1.2
ISTIO_VERSION = 1.16.1
BUF_VERSION = 1.7.0

PATH := $(PATH):$(PWD)/bin

all: check manager

.PHONY: check
check: fmt vet test lint ## Run tests and linters

# Check that all generated code was checked in to git
check-all-code-generation: check-generate check-manifests

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b ./bin v${GOLANGCI_VERSION}
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint ## Run linter
# "unused" linter is a memory hog, but running it separately keeps it contained (probably because of caching)
	bin/golangci-lint run --disable=unused -c .golangci.yml --timeout 2m
	bin/golangci-lint run -c .golangci.yml --timeout 2m

bin/licensei: bin/licensei-${LICENSEI_VERSION}
	@ln -sf licensei-${LICENSEI_VERSION} bin/licensei
bin/licensei-${LICENSEI_VERSION}:
	@mkdir -p bin
	curl -sfL https://raw.githubusercontent.com/goph/licensei/master/install.sh | bash -s v${LICENSEI_VERSION}
	@mv bin/licensei $@

.PHONY: license-check
license-check: bin/licensei ## Run license check
	bin/licensei check
	bin/licensei header

.PHONY: license-cache
license-cache: bin/licensei ## Generate license cache
	bin/licensei cache

.PHONY: vendor-clean
vendor-clean: ## Clean vendor cache
	rm -rf ${PWD}/vendor

# Run tests
.PHONY: test
test: install-envtest
ifneq (${SKIP_TESTS}, 1)
	KUBEBUILDER_ASSETS="$${PWD}/bin/envtest/bin" go test ./... -coverprofile cover.out
endif

# Build manager binary
.PHONY: manager
manager: generate manifests fmt vet build

# Build manager binary
.PHONY: build
build:
	go build -ldflags="-X main.SupportedIstioVersion=${ISTIO_VERSION} -X main.Version=${TAG}" -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install kustomize
install-kustomize:
	scripts/install_kustomize.sh ${KUSTOMIZE_VERSION}

# Install envtest
install-envtest:
	scripts/install_envtest.sh ${ENVTEST_K8S_VERSION}

# Install CRDs into a cluster
install: install-kustomize manifests
	bin/kustomize build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: install-kustomize manifests
	bin/kustomize build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: install-kustomize manifests
	cd config/manager && ../../bin/kustomize edit set image controller=${IMG}
	bin/kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC, OpenAPI etc.
manifests: download-deps update-istio-deps
	bin/controller-gen rbac:roleName=manager-role webhook paths="./..."
	bin/cue-gen -paths=build -f=cue.yaml
	bin/cue-gen -paths=build -f=cue.yaml -crd
	./scripts/label-crds.sh $(ISTIO_VERSION)
	cp -a config/crd/bases/* deploy/charts/istio-operator/crds

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Download build dependencies
download-deps:
ifneq (${SKIP_UPDATE_DEPS}, 1)
	./scripts/download-deps.sh
	./scripts/install-buf.sh $(BUF_VERSION)
endif

# Update Istio build dependencies
update-istio-deps:
ifneq (${SKIP_UPDATE_DEPS}, 1)
	./scripts/update-istio-dependencies.sh $(ISTIO_VERSION)
endif

# Generate code
generate: download-deps update-istio-deps
	cd build && ../bin/buf generate --path api
	go run ./build/fixup_structs/main.go -f api/v1alpha1/common.pb.go
	cd api/v1alpha1 && ../../bin/controller-gen object:headerFile="../../hack/boilerplate.go.txt" paths="./..."

# Check that code generation was checked in to git
check-generate: generate
	git diff --exit-code

# Check that manifests were checked in to git
check-manifests: manifests
	git diff --exit-code

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

check_release:
	@echo "New tags (${REL_TAG}, ${API_REL_TAG}, ${EMBEDDED_CHART_REL_TAG} and ${HELM_CHART_REL_TAG}) will be pushed to Github, a new Docker image (${REL_TAG}) will be released, and a new Helm chart (${HELM_CHART_REL_TAG}) will be released. Are you sure? [y/N] " && read -r ans && [ "$${ans:-N}" = y ]

release: check_release
	git tag -a ${REL_TAG} -m ${RELEASE_MSG}
	git tag -a ${API_REL_TAG} -m ${API_RELEASE_MSG}
	git tag -a ${EMBEDDED_CHART_REL_TAG} -m ${EMBEDDED_CHART_RELEASE_MSG}
	git tag -a ${HELM_CHART_REL_TAG} -m ${HELM_CHART_RELEASE_MSG}
	git push origin ${REL_TAG}
	git push origin ${API_REL_TAG}
	git push origin ${EMBEDDED_CHART_REL_TAG}
	git push origin ${HELM_CHART_REL_TAG}

.PHONY: tidy
tidy: ## Execute go mod tidy
	go mod tidy
	go mod download all

#Generate CRD docs
.PHONY: docs
docs:
	go run cmd/docs.go
