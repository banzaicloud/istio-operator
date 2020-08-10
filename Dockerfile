ARG GO_VERSION=1.13.9

# Build the manager binary
FROM golang:${GO_VERSION}-alpine3.11 AS builder

RUN apk add --update --no-cache ca-certificates make~=4.2 git~=2.24 curl~=7.67 mercurial~=5.3

ARG PACKAGE=github.com/banzaicloud/istio-operator

RUN mkdir -p /go/src/${PACKAGE}
WORKDIR /go/src/${PACKAGE}

COPY go.mod go.sum ./
RUN go mod download

COPY pkg/    pkg/
COPY cmd/    cmd/
COPY Makefile go.* /go/src/${PACKAGE}/
RUN make vendor

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/banzaicloud/istio-operator/cmd/manager

# Copy the controller-manager into a thin image
FROM alpine:3.11
RUN apk add --no-cache ca-certificates
WORKDIR /
COPY --from=builder /go/src/github.com/banzaicloud/istio-operator/manager .
ENTRYPOINT ["/manager"]
