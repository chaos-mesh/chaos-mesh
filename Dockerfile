# syntax=docker/dockerfile:experimental

FROM debian:buster-slim AS go_build

ENV DEBIAN_FRONTEND noninteractive

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install build-essential curl git pkg-config libfuse-dev fuse -y

RUN curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
RUN echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list

RUN apt-get update && apt-get install yarn -y

RUN curl https://dl.google.com/go/go1.14.6.linux-amd64.tar.gz | tar -xz -C /usr/local
ENV PATH "/usr/local/go/bin:${PATH}"
ENV GO111MODULE=on

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN if [[ -n "$HTTP_PROXY" ]]; then yarn config set proxy $HTTP_PROXY; fi

WORKDIR /src

COPY . /src

ARG UI
ARG SWAGGER
ARG LDFLAGS

RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/src/ui/node_modules \
    IMG_LDFLAGS=$LDFLAGS make binary

WORKDIR /bin

RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.1.9/toda-linux-amd64.tar.gz | tar -xz
RUN curl -L https://github.com/chaos-mesh/nsexec/releases/download/v0.1.5/nsexec-linux-amd64.tar.gz | tar -xz

RUN cp /src/bin/* /bin
