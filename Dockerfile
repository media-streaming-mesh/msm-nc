# Build the manager binary
FROM --platform=$BUILDPLATFORM golang:1.20.5 as builder

WORKDIR /workspace

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY api/ api/
COPY internal/controller/ internal/controller/

ARG TARGETOS TARGETARCH

# Build
RUN --mount=type=cache,target=/root/.cache/go-build \
        --mount=type=cache,target=/go/pkg \
        GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -o msm-network-controller cmd/msm-nc/main.go
# TODO above point to correct main.go
# TODO remove below before final branch merge
# Kubebuilder build line
#RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#FROM gcr.io/distroless/static:nonroot
FROM ubuntu
WORKDIR /
COPY --from=builder /workspace/msm-network-controller .
#USER nonroot:nonroot

# need to fix this up
ENTRYPOINT ["/msm-network-controller"]
