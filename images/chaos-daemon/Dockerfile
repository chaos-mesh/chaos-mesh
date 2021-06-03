FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux procps curl && rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE 1

RUN curl -L https://github.com/chaos-mesh/toda/releases/download/v0.2.0/toda-linux-amd64.tar.gz -o /usr/local/bin/toda.tar.gz
RUN curl -L https://github.com/chaos-mesh/nsexec/releases/download/v0.1.5/nsexec-linux-amd64.tar.gz -o /usr/local/bin/nsexec.tar.gz
RUN curl -L https://github.com/chaos-mesh/rs-tproxy/releases/download/v0.2.1/tproxy-linux-amd64.tar.gz -o /usr/local/bin/tproxy.tar.gz

RUN tar xvf /usr/local/bin/toda.tar.gz -C /usr/local/bin
RUN rm /usr/local/bin/toda.tar.gz
RUN tar xvf /usr/local/bin/tproxy.tar.gz -C /usr/local/bin
RUN rm /usr/local/bin/tproxy.tar.gz
RUN tar xvf /usr/local/bin/nsexec.tar.gz -C /usr/local/bin
RUN rm /usr/local/bin/nsexec.tar.gz
RUN cp /usr/local/bin/libnsenter.so /usr/local/lib/libnsenter.so

COPY bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY bin/pause /usr/local/bin/pause
