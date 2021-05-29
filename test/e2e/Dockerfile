# This is for running the tests in a docker container which is useful on non-linux systems where MetalLB
# in effectively non-functional

ARG GO_VERSION=1.13.9

# Build the manager binary
FROM golang:${GO_VERSION}-alpine3.11

RUN apk add --update --no-cache make~=4.2 git~=2.24 bash~=5 gcc~=9.3 musl-dev

ENV PACKAGE=github.com/banzaicloud/istio-operator
ENV WORKDIR=/go/src/${PACKAGE}
RUN mkdir -p ${WORKDIR}
WORKDIR ${WORKDIR}

ENV E2E_TEST_FAIL_FAST=0
ENV E2E_LOG_DIR=/logs
ENV E2E_TEST_GINKGO_ARGS=""

VOLUME ${E2E_LOG_DIR}

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/    pkg/
COPY cmd/    cmd/
COPY Makefile ./
COPY scripts/ scripts/
COPY test/ test/
RUN make vendor

RUN make download-deps

CMD make e2e-test-run
