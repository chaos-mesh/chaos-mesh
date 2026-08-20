FROM curlimages/curl:7.88.1 AS binaryimage

USER root

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

ARG TARGET_PLATFORM=amd64

RUN mkdir -p /tmp/bin && \
    curl -L https://mirrors.chaos-mesh.org/byteman-chaos-mesh-download-v4.0.24-0.12.tar.gz -o /usr/local/byteman.tar.gz && \
    tar xvf /usr/local/byteman.tar.gz -C /usr/local && \
    mv /usr/local/byteman-chaos-mesh-download-v4.0.24-0.12 /tmp/byteman && \
    rm /usr/local/byteman.tar.gz

# toda doesn't support arm64 yet
RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.2.4/toda-linux-amd64.tar.gz | tar xz -C /tmp/bin
RUN case "$TARGET_PLATFORM" in \
    'amd64') \
    export NSEXEC_ARCH='x86_64'; \
    ;; \
    'arm64') \
    export NSEXEC_ARCH='aarch64'; \
    ;; \
    *) echo >&2 "error: unsupported architecture '$TARGET_PLATFORM'"; exit 1 ;; \
    esac; \
    curl -L https://github.com/chaos-mesh/nsexec/releases/download/v0.1.6/nsexec-$NSEXEC_ARCH-unknown-linux-gnu.tar.gz | tar xz -C /tmp/bin; \
    curl -L https://github.com/chaos-mesh/chaos-tproxy/releases/download/v0.5.3/tproxy-$NSEXEC_ARCH.tar.gz | tar xz -C /tmp/bin; \
    curl -L https://github.com/chaos-mesh/memStress/releases/download/v0.3.1/memStress_v0.3.1-$NSEXEC_ARCH-linux-gnu.tar.gz | tar xz -C /tmp/bin

# ---
FROM alpine:3.23 AS shimbuilder
RUN apk add --no-cache gcc musl-dev
RUN echo 'int __res_init(void) { return 0; } int res_init(void) { return 0; }' > res_init.c
RUN gcc -shared -fPIC -o libresinit.so res_init.c

# ---
FROM rust:alpine3.23 AS rustbuilder
RUN apk add --no-cache git musl-dev gcc
RUN git clone --depth 1 --branch v0.1.6 https://github.com/chaos-mesh/nsexec.git /tmp/nsexec && \
    cd /tmp/nsexec && \
    rm -f rust-toolchain && \
    rm -rf .cargo && \
    RUSTFLAGS="-C target-feature=-crt-static" cargo build --release --all

# ---
FROM alpine:3.23

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

RUN apk upgrade --no-cache && \
    apk add --no-cache \
        tzdata \
        iptables-legacy \
        ebtables \
        ipset \
        stress-ng \
        iproute2 \
        fuse \
        util-linux \
        procps \
        openjdk17-jre-headless \
        gcompat \
        libgcc

RUN ln -sf /sbin/iptables-legacy /sbin/iptables && \
    ln -sf /sbin/iptables-legacy-save /sbin/iptables-save && \
    ln -sf /sbin/ip6tables-legacy /sbin/ip6tables && \
    ln -sf /sbin/ip6tables-legacy-save /sbin/ip6tables-save && \
    ln -sf /sbin/ebtables-legacy /sbin/ebtables

ENV RUST_BACKTRACE=1
ENV BYTEMAN_HOME=/usr/local/byteman
ENV PATH=$PATH:/usr/local/byteman/bin
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk
ENV PATH=$PATH:/usr/local/bin

COPY bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY bin/pause /usr/local/bin/pause
COPY bin/cdh /usr/local/bin/cdh
COPY --from=shimbuilder libresinit.so /usr/local/lib/libresinit.so
COPY --from=binaryimage /tmp/byteman /usr/local/byteman
COPY --from=binaryimage /tmp/bin/* /usr/local/bin/

COPY --from=rustbuilder /tmp/nsexec/target/release/nsexec /usr/local/bin/nsexec
COPY --from=rustbuilder /tmp/nsexec/target/release/libnsenter.so /usr/local/lib/libnsenter.so

RUN rm -f /usr/local/bin/libnsenter.so && \
    mv /usr/local/bin/tproxy /usr/local/bin/tproxy-bin && \
    printf '#!/bin/sh\nexport LD_PRELOAD=/usr/local/lib/libresinit.so\nexec /usr/local/bin/tproxy-bin "$@"\n' > /usr/local/bin/tproxy && \
    chmod +x /usr/local/bin/tproxy
