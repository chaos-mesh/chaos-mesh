# syntax=docker/dockerfile:1

FROM golang:1.27.3 AS go_install

ENV GOPATH=/go
ENV CGO_ENABLED=0

RUN go install golang.org/x/tools/cmd/goimports@v0.52.0
RUN go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.22.0
RUN go install github.com/mgechev/revive@v1.17.1
RUN go install github.com/pingcap/failpoint/failpoint-ctl@v0.0.0-20200210140405-f8f9fb234798
# gosec's pinned x/tools cannot read Go 1.27 export data.
RUN mkdir /gosec-build && cd /gosec-build && \
    go mod init gosec-build && \
    go get github.com/securego/gosec/v2/cmd/gosec@v2.29.0 golang.org/x/tools@v0.52.0 && \
    go install github.com/securego/gosec/v2/cmd/gosec && \
    rm -rf /gosec-build
RUN go install github.com/golang/protobuf/protoc-gen-go@v1.5.4
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6
RUN go install github.com/onsi/ginkgo/v2/ginkgo@v2.32.0

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

ARG TARGETARCH

RUN --mount=type=cache,id=chaos-mesh-debian-bookworm-apt-${TARGETARCH},target=/var/cache/apt,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get -o Acquire::Retries=3 update && \
    apt-get -o Acquire::Retries=3 install unzip git build-essential curl musl musl-dev ebtables -y && \
    rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set ebtables /usr/sbin/ebtables-legacy

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
    curl -L https://github.com/protocolbuffers/protobuf/releases/download/v36.2/protoc-36.2-linux-$PROTOC_ARCH.zip > /protoc.zip && \
    unzip /protoc.zip -d /usr/local && \
    rm /protoc.zip
RUN chmod +rx -R /usr/local/include
RUN chmod +x /usr/local/bin/protoc

RUN mkdir -p /go/bin

COPY --from=go_install /usr/local/go /usr/local/go
COPY --from=go_install /go/bin /go/bin

ENV PATH="/usr/local/go/bin:${PATH}:/tmp/go/bin:/go/bin"

RUN curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v5.8.3/kustomize_v5.8.3_$(go env GOOS)_$(go env GOARCH).tar.gz | tar -xz -C /usr/local/bin/

# The folder structure of envtest is: controller-tools/envtest/*.
RUN curl -L https://github.com/kubernetes-sigs/controller-tools/releases/download/envtest-v1.37.0/envtest-v1.37.0-$(go env GOOS)-$(go env GOARCH).tar.gz | tar -xz -C /usr/local/
ENV KUBEBUILDER_ASSETS=/usr/local/controller-tools/envtest

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/v4.3.1/scripts/get-helm-4 && \
    chmod 700 get_helm.sh && \
    ./get_helm.sh --version v4.3.1 && \
    rm get_helm.sh

# Create and grant permissions to plugins directory.
RUN mkdir -p /opt/helm/plugins && chmod a+rwx /opt/helm
ENV HELM_PLUGINS=/opt/helm/plugins
RUN helm plugin install https://github.com/losisin/helm-values-schema-json.git --version v2.6.0 --verify=false

RUN mkdir /.cache
RUN chmod -R 777 /.cache

ENV GOCACHE=/tmp/go-build
ENV GOPATH=/tmp/go
ENV CGO_ENABLED=0

LABEL org.opencontainers.image.source=https://github.com/chaos-mesh/chaos-mesh
