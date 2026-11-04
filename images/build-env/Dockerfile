# syntax=docker/dockerfile:1

FROM node:24.22.1-bookworm-slim AS node-runtime

FROM golang:1.27.3-bookworm AS go-runtime

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

ARG TARGETARCH
ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN --mount=type=cache,id=chaos-mesh-debian-bookworm-apt-${TARGETARCH},target=/var/cache/apt,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean && \
    apt-get -o Acquire::Retries=3 update && \
    apt-get -o Acquire::Retries=3 install -y build-essential ca-certificates curl git pkg-config libfuse-dev fuse libatomic1 python3 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=go-runtime /usr/local/go /usr/local/go
COPY --from=node-runtime /usr/local/bin/node /usr/local/bin/node
COPY --from=node-runtime /usr/local/lib/node_modules /usr/local/lib/node_modules
RUN ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm && \
    npm install -g pnpm@12.14.0
ENV GOLANG_VERSION=1.27.3
ENV GOTOOLCHAIN=local
ENV PATH=/go/bin:/usr/local/go/bin:$PATH
RUN mkdir -p /go/src /go/bin && chmod -R 1777 /go
WORKDIR /go

ENV GOCACHE=/tmp/go-build
ENV GOPATH=/tmp/go

LABEL org.opencontainers.image.source=https://github.com/chaos-mesh/chaos-mesh
