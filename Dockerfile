ARG GID=1000
ARG UID=1000

# Build the manager binary
FROM golang:1.18 as builder
ARG GID
ARG UID

# Create user and group
RUN groupadd -g ${GID} appgroup && \
    useradd -u ${UID} --gid appgroup appuser

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# Copy the API Go Modules manifests
COPY api/go.mod api/go.mod
COPY api/go.sum api/go.sum
# Copy the deploy/charts Go Modules manifests
COPY deploy/charts/go.mod deploy/charts/go.mod
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY deploy/ deploy/
COPY internal/ internal/
COPY pkg/ pkg/
COPY Makefile Makefile

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
ARG GID
ARG UID

WORKDIR /
COPY --from=builder /workspace/bin/manager /manager

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
USER ${UID}:${GID}

ENTRYPOINT ["/manager"]
