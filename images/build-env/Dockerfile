# syntax=docker/dockerfile:experimental

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN apt update && apt install -y build-essential curl git pkg-config libfuse-dev fuse
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt install -y nodejs
RUN rm -rf /var/lib/apt/lists/*
RUN npm install -g pnpm@9

ARG TARGET_PLATFORM=amd64

RUN curl -fsS https://dl.google.com/go/go1.24.11.linux-$TARGET_PLATFORM.tar.gz | tar -xz -C /usr/local
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOCACHE=/tmp/go-build
ENV GOPATH=/tmp/go

LABEL org.opencontainers.image.source=https://github.com/chaos-mesh/chaos-mesh
