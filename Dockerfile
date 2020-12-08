# syntax=docker/dockerfile:experimental

FROM pingcap/chaos-build-base AS go_build

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

FROM alpine:3.12

RUN apk add --no-cache curl tar

WORKDIR /bin

RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.1.9/toda-linux-amd64.tar.gz | tar -xz
RUN curl -L https://github.com/YangKeao/nsexec/releases/download/v0.1.5/nsexec-linux-amd64.tar.gz | tar -xz

COPY ./scripts /scripts
COPY --from=go_build /src/bin /bin
