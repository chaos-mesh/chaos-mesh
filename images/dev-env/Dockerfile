# syntax=docker/dockerfile:experimental

FROM golang:1.20 AS go_install

ENV GOPATH /go
ENV CGO_ENABLED 0

RUN go install golang.org/x/tools/cmd/goimports@v0.13.0
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
RUN go install github.com/mgechev/revive@v1.2.4
RUN go install github.com/pingcap/failpoint/failpoint-ctl@v0.0.0-20200210140405-f8f9fb234798
RUN go install github.com/securego/gosec/cmd/gosec@v0.0.0-20200401082031-e946c8c39989
RUN go install github.com/99designs/gqlgen@v0.17.2
RUN go install github.com/golang/protobuf/protoc-gen-go@v1.4.2
RUN go install github.com/axw/gocov/gocov@v1.0.0
RUN go install github.com/AlekSi/gocov-xml@v0.0.0-20190121064608-3a14fb1c4737
RUN go install github.com/matm/gocov-html@v0.0.0-20200509184451-71874e2e203b
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.7
RUN go install github.com/onsi/ginkgo/v2/ginkgo@v2.12.0
RUN go install github.com/apache/skywalking-eyes/cmd/license-eye@v0.2.0

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND noninteractive

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt update && \
    apt install unzip git build-essential curl musl musl-dev -y && \
    rm -rf /var/lib/apt/lists/*

# The `TARGET_PLATFORM` would be `amd64` or `arm64`
ARG TARGET_PLATFORM=amd64

# The architecture part of the url is `aarch_64` or `x86_64`
RUN case "$TARGET_PLATFORM" in \
    'amd64') \
    export PROTOC_ARCH='x86_64'; \
    ;; \
    'arm64') \
    export PROTOC_ARCH='aarch_64'; \
    ;; \
    *) echo >&2 "error: unsupported architecture '$TARGET_PLATFORM'"; exit 1 ;; \
    esac; \
    curl -L https://github.com/protocolbuffers/protobuf/releases/download/v3.12.2/protoc-3.12.2-linux-$PROTOC_ARCH.zip > /protoc.zip && \
    unzip /protoc.zip -d /usr/local && \
    rm /protoc.zip
RUN chmod +rx -R /usr/local/include
RUN chmod +x /usr/local/bin/protoc

RUN mkdir -p /go/bin

COPY --from=go_install /usr/local/go /usr/local/go
COPY --from=go_install /go/bin /go/bin

# The `gqlgen` depends on the plugin in it
COPY --from=go_install /go/pkg/mod/github.com/99designs/gqlgen@v0.17.2 /go/pkg/mod/github.com/99designs/gqlgen@v0.17.2

ENV PATH "/usr/local/go/bin:${PATH}:/tmp/go/bin:/go/bin"

RUN curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v4.3.0/kustomize_v4.3.0_$(go env GOOS)_$(go env GOARCH).tar.gz | tar -xz -C /usr/local/bin/

RUN curl -L https://storage.googleapis.com/kubebuilder-tools/kubebuilder-tools-1.19.2-$(go env GOOS)-$(go env GOARCH).tar.gz | tar -xz -C /usr/local/

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && \
    chmod 700 get_helm.sh && \
    ./get_helm.sh && \
    rm get_helm.sh

# Create and grant permissions to plugins directory.
RUN mkdir -p /opt/helm/plugins && chmod a+rwx /opt/helm
ENV HELM_PLUGINS=/opt/helm/plugins
RUN helm plugin install https://github.com/losisin/helm-values-schema-json.git

RUN mkdir /.cache
RUN chmod -R 777 /.cache

ENV GOCACHE /tmp/go-build
ENV GOPATH /tmp/go
ENV CGO_ENABLED 0

LABEL org.opencontainers.image.source https://github.com/chaos-mesh/chaos-mesh
