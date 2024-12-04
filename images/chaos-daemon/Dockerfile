FROM curlimages/curl:7.88.1 as binaryimage

USER root

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

ARG TARGET_PLATFORM=amd64

RUN mkdir -p /tmp/bin && \
    curl -L https://mirrors.chaos-mesh.org/byteman-chaos-mesh-download-v4.0.22-0.12.tar.gz -o /usr/local/byteman.tar.gz && \
    tar xvf /usr/local/byteman.tar.gz -C /usr/local && \
    mv /usr/local/byteman-chaos-mesh-download-v4.0.22-0.12 /tmp/byteman && \
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
    curl -L https://github.com/chaos-mesh/memStress/releases/download/v0.3/memStress_v0.3-$NSEXEC_ARCH-linux-gnu.tar.gz | tar xz -C /tmp/bin




# ---
FROM debian:bullseye-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN mkdir -p /usr/share/man/man1 /usr/share/man/man2

RUN apt-get update && \
    echo "deb http://security.debian.org/debian-security bullseye-security main contrib non-free" >> /etc/apt/sources.list && \
    echo "deb http://deb.debian.org/debian bullseye-proposed-updates main contrib non-free" >> /etc/apt/sources.list && \
    apt --allow-releaseinfo-change update && apt upgrade -y && \
    apt-get install -y tzdata iptables ebtables net-tools ipset stress-ng iproute2 fuse util-linux procps openjdk-11-jre && \
    rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy && \
    update-alternatives --set ebtables /usr/sbin/ebtables-legacy

ENV RUST_BACKTRACE 1
ENV BYTEMAN_HOME /usr/local/byteman
ENV PATH $PATH:/usr/local/byteman/bin
ENV JAVA_HOME /usr/lib/jvm/java-11-openjdk-amd64
ENV PATH $PATH:/usr/local/bin

COPY bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY bin/pause /usr/local/bin/pause
COPY bin/cdh /usr/local/bin/cdh
COPY --from=binaryimage /tmp/byteman /usr/local/byteman
COPY --from=binaryimage /tmp/bin/* /usr/local/bin/
RUN mv /usr/local/bin/libnsenter.so /usr/local/lib/libnsenter.so
