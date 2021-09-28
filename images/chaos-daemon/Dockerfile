FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && \ 
    apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux procps curl && \
    apt-get clean

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE 1

# toda doesn't support arm64 yet
RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.2.0/toda-linux-amd64.tar.gz | tar -xz -C /usr/local/bin

RUN curl -L https://github.com/chaos-mesh/nsexec/releases/download/v0.1.6/nsexec-$(uname -m)-unknown-linux-gnu.tar.gz | tar -xz -C /usr/local/bin &&\
    mv /usr/local/bin/libnsenter.so /usr/local/lib/libnsenter.so
RUN curl -L https://github.com/chaos-mesh/rs-tproxy/releases/download/v0.2.4/tproxy-$(uname -m)-unknown-linux-gnu.tar.gz | tar -xz -C /usr/local/bin

COPY bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY bin/pause /usr/local/bin/pause
