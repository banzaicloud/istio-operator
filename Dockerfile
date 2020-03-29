ARG GO_VERSION=1.13.9

# Build the manager binary
FROM golang:${GO_VERSION}-alpine3.11 AS builder

RUN apk add --update --no-cache ca-certificates=20191127-r1 make=4.2.1-r2 git=2.24.1-r0 curl=7.67.0-r0 mercurial=5.3.1-r0

ARG PACKAGE=github.com/banzaicloud/istio-operator

RUN mkdir -p /go/src/${PACKAGE}
WORKDIR /go/src/${PACKAGE}

COPY pkg/    pkg/
COPY cmd/    cmd/
COPY Makefile Gopkg.* /go/src/${PACKAGE}/
COPY vendor/ vendor/
RUN make vendor

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/banzaicloud/istio-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.11
RUN apk add --no-cache ca-certificates=20191127-r1
WORKDIR /
COPY --from=builder /go/src/github.com/banzaicloud/istio-operator/manager .
ENTRYPOINT ["/manager"]
