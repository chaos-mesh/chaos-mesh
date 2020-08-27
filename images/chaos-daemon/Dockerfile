FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apt-get update
RUN apt-get install -y tzdata iptables ipset stress-ng iproute2 fuse util-linux

ENV RUST_BACKTRACE 1

COPY --from=pingcap/binary /bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY --from=pingcap/binary /bin/toda /usr/local/bin/toda
COPY --from=pingcap/binary /bin/pause /usr/local/bin/pause
COPY --from=pingcap/binary /bin/suicide /usr/local/bin/suicide
