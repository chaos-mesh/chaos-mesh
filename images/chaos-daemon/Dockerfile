FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && \ 
    apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux procps curl && \
    rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE 1
ARG TARGET_PLATFORM=amd64

# toda doesn't support arm64 yet
RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.2.0/toda-linux-amd64.tar.gz | tar xz -C /usr/local/bin
RUN case "$TARGET_PLATFORM" in \
    'amd64') \
    export NSEXEC_ARCH='x86_64'; \
    ;; \
    'arm64') \
    export NSEXEC_ARCH='aarch64'; \
    ;; \
    *) echo >&2 "error: unsupported architecture '$TARGET_PLATFORM'"; exit 1 ;; \
    esac; \
    curl -L https://github.com/chaos-mesh/nsexec/releases/download/v0.1.6/nsexec-$NSEXEC_ARCH-unknown-linux-gnu.tar.gz | tar xz -C /usr/local/bin; \
    mv /usr/local/bin/libnsenter.so /usr/local/lib/libnsenter.so; \
    curl -L https://github.com/chaos-mesh/rs-tproxy/releases/download/v0.2.4/tproxy-$NSEXEC_ARCH-unknown-linux-gnu.tar.gz | tar xz -C /usr/local/bin

COPY bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY bin/pause /usr/local/bin/pause
