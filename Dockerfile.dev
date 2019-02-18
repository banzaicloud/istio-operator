ARG GO_VERSION=1.11

# Build the manager binary
FROM golang:${GO_VERSION}-alpine AS builder

RUN apk add --update --no-cache ca-certificates make git curl mercurial

ARG PACKAGE=github.com/banzaicloud/istio-operator

RUN mkdir -p /go/src/${PACKAGE}
WORKDIR /go/src/${PACKAGE}

COPY pkg/    pkg/
COPY cmd/    cmd/
COPY vendor/ vendor/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/banzaicloud/istio-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.7
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/banzaicloud/istio-operator/manager .
ENTRYPOINT ["/manager"]
