ARG GO_VERSION=1.15.15

# Build the manager binary
FROM golang:${GO_VERSION}-alpine3.13 AS builder

# hadolint ignore=DL3018
RUN apk add --update --no-cache ca-certificates make git curl mercurial bash

ARG PACKAGE=github.com/banzaicloud/istio-operator

RUN mkdir -p /go/src/${PACKAGE}
WORKDIR /go/src/${PACKAGE}

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/    pkg/
COPY cmd/    cmd/
COPY test/e2e/e2e-test.mk /go/src/${PACKAGE}/test/e2e/
COPY Makefile go.* /go/src/${PACKAGE}/
RUN make vendor

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/banzaicloud/istio-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.13.5
RUN apk add --no-cache ca-certificates=20191127-r5
WORKDIR /
COPY --from=builder /go/src/github.com/banzaicloud/istio-operator/manager .
ENTRYPOINT ["/manager"]
