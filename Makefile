# Image URL to use all building/pushing image targets
TAG ?= $(shell git describe --tags --abbrev=0 --match '[0-9].*[0-9].*[0-9]' 2>/dev/null )
IMG ?= banzaicloud/istio-operator:$(TAG)

RELEASE_TYPE ?= p
RELEASE_MSG ?= "operator release"

REL_TAG = $(shell ./scripts/increment_version.sh -${RELEASE_TYPE} ${TAG})

GOLANGCI_VERSION = 1.23.8
LICENSEI_VERSION = 0.1.0

KUSTOMIZE_BASE = config/overlays/specific-manager-version

all: test manager

.PHONY: check
check: test lint shellcheck-makefile shellcheck ## Run tests and linters

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
test: install-kubebuilder generate fmt vet manifests
	KUBEBUILDER_ASSETS="$${PWD}/bin/kubebuilder/bin" go test ./pkg/... ./cmd/... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager github.com/banzaicloud/istio-operator/cmd/manager

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./cmd/manager/main.go

# Install kustomize
install-kustomize: bin/kustomize
bin/kustomize:
	scripts/install_kustomize.sh

# Install kubebuilder
install-kubebuilder: bin/kubebuilder/bin/kubebuilder
bin/kubebuilder/bin/kubebuilder:
	scripts/install_kubebuilder.sh

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
manifests:
	bin/controller-gen rbac --output-dir config/base/rbac
	bin/controller-gen crd --output-dir config/base/crds
	find config/base/crds -exec touch -t 201901010101 {} +

# Run go fmt against code
fmt:
	go fmt ./pkg/... ./cmd/...

# Run go vet against code
vet:
	go vet ./pkg/... ./cmd/...

download-deps:
	./scripts/download-deps.sh

# Generate code
generate: download-deps
	find config/base/crds -exec touch -t 201901010101 {} +
	find pkg/manifests/istio-crds -exec touch -t 201901010101 {} +
	go generate ./pkg/... ./cmd/...
	./hack/update-codegen.sh

# Verify codegen
verify-codegen: download-deps
	./hack/verify-codegen.sh

# Build the docker image
docker-build:
	docker build -f Dockerfile.dev . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}

check_release:
	@echo "A new tag (${REL_TAG}) will be pushed to Github, and a new Docker image will be released. Are you sure? [y/N] " && read -r ans && [ "$${ans:-N}" = y ]

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
