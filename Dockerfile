ARG GO_VERSION=1.16.4

# Build the manager binary
FROM golang:${GO_VERSION}-alpine3.13 AS builder

RUN apk add --update --no-cache ca-certificates make~=4.3 git~=2.30 curl~=7.77 mercurial~=5.5 bash~=5

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
FROM alpine:3.13
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/banzaicloud/istio-operator/manager .
ENTRYPOINT ["/manager"]
