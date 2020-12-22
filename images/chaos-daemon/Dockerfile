FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux procps && rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE 1

COPY bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY bin/pause /usr/local/bin/pause

COPY bin/toda.tar.gz /usr/local/bin/toda.tar.gz
RUN tar xvf /usr/local/bin/toda.tar.gz -C /usr/local/bin
RUN rm /usr/local/bin/toda.tar.gz

COPY bin/nsexec.tar.gz /usr/local/bin/nsexec.tar.gz
RUN tar xvf /usr/local/bin/nsexec.tar.gz -C /usr/local/bin
RUN rm /usr/local/bin/nsexec.tar.gz
