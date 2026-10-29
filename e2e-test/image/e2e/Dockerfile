FROM alpine:3.23

ARG TARGET_PLATFORM=amd64

ENV KUBECTL_VERSION=v1.34.12
ENV HELM_VERSION=v4.3.1

RUN apk update && apk add --no-cache ca-certificates curl git openssl bash
RUN curl -fsSL https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${TARGET_PLATFORM}/kubectl \
    -o /usr/local/bin/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    curl -fsSL https://get.helm.sh/helm-${HELM_VERSION}-linux-${TARGET_PLATFORM}.tar.gz \
    -o helm-${HELM_VERSION}-linux-${TARGET_PLATFORM}.tar.gz && \
    tar -zxvf helm-${HELM_VERSION}-linux-${TARGET_PLATFORM}.tar.gz && \
    mv linux-${TARGET_PLATFORM}/helm /usr/local/bin/helm && \
    rm -rf linux-${TARGET_PLATFORM} && \
    rm helm-${HELM_VERSION}-linux-${TARGET_PLATFORM}.tar.gz


ADD chaos-mesh /charts/e2e/chaos-mesh
ADD manifests /manifests/e2e

ADD bin/ginkgo /usr/local/bin/
ADD bin/e2e.test /usr/local/bin/
