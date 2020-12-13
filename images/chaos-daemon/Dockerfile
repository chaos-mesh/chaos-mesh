FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux procps && rm -rf /var/lib/apt/lists/*

RUN update-alternatives --set iptables /usr/sbin/iptables-legacy

ENV RUST_BACKTRACE 1

COPY --from=pingcap/chaos-binary /bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY --from=pingcap/chaos-binary /bin/toda /usr/local/bin/toda
COPY --from=pingcap/chaos-binary /bin/pause /usr/local/bin/pause
COPY --from=pingcap/chaos-binary /bin/suicide /usr/local/bin/suicide
COPY --from=pingcap/chaos-binary /toda-glibc /usr/local/lib/toda-glibc
COPY --from=pingcap/chaos-binary /bin/nsexec /usr/local/bin/nsexec
COPY --from=pingcap/chaos-binary /bin/libnsenter.so /usr/local/lib/libnsenter.so
