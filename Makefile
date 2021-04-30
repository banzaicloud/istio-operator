# Image URL to use all building/pushing image targets
TAG ?= $(shell git describe --tags --abbrev=0 --match '[0-9].*[0-9].*[0-9]' 2>/dev/null )
IMG ?= banzaicloud/istio-operator-v2:$(TAG)

GOLANGCI_VERSION = 1.39.0
LICENSEI_VERSION = 0.3.1
KUBEBUILDER_VERSION = 2.3.2
KUSTOMIZE_VERSION = 4.1.2
ISTIO_VERSION = 1.10.0-rc.0
SKIP_ISTIO_API_UPDATE = false
BUF_VERSION = 0.41.0

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
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s -- -b ./bin v${GOLANGCI_VERSION}
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

# Run tests
.PHONY: test
test: install-kubebuilder
	KUBEBUILDER_ASSETS="$${PWD}/bin/kubebuilder/bin" go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet build

# Build manager binary
build:
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Install kustomize
install-kustomize:
	scripts/install_kustomize.sh ${KUSTOMIZE_VERSION}

# Install kubebuilder
install-kubebuilder:
	scripts/install_kubebuilder.sh ${KUBEBUILDER_VERSION}

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

# Generate manifests e.g. CRD, RBAC etc.
manifests: download-deps update-istio-deps
	bin/controller-gen rbac:roleName=manager-role webhook paths="./..."
	bin/cue-gen -paths=build -f=cue.yaml -crd

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Download build dependencies
download-deps:
	./scripts/download-deps.sh
	./scripts/install-buf.sh $(BUF_VERSION)

# Update Istio build dependencies
update-istio-deps:
	./scripts/update-istio-dependencies.sh $(ISTIO_VERSION) $(SKIP_ISTIO_API_UPDATE)

# Generate code
generate: download-deps update-istio-deps
	cd build && ../bin/buf generate --path api
	bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

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
