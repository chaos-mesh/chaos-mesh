# syntax=docker/dockerfile:experimental

FROM debian:buster-slim

ENV DEBIAN_FRONTEND noninteractive

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install build-essential curl git pkg-config libfuse-dev fuse -y

# require nodejs >= 12
RUN curl -fsSL https://deb.nodesource.com/setup_14.x | bash -
RUN apt-get install -y nodejs
RUN npm config set unsafe-perm true
RUN npm install yarn -g

RUN if [[ -n "$HTTP_PROXY" ]]; then yarn config set proxy $HTTP_PROXY; fi

ARG TARGET_PLATFORM=amd64
RUN curl https://dl.google.com/go/go1.16.7.linux-$TARGET_PLATFORM.tar.gz | tar -xz -C /usr/local
ENV PATH "/usr/local/go/bin:${PATH}"
ENV GO111MODULE=on
ENV GOCACHE /tmp/go-build
ENV GOPATH /tmp/go

LABEL org.opencontainers.image.source https://github.com/chaos-mesh/chaos-mesh