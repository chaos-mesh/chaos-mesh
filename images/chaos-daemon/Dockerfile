FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add --no-cache tzdata iptables ipset stress-ng iproute2 util-linux

COPY --from=pingcap/binary /src/bin/chaos-daemon /usr/local/bin/chaos-daemon
COPY --from=pingcap/binary /src/bin/pause /usr/local/bin/pause
COPY --from=pingcap/binary /src/bin/suicide /usr/local/bin/suicide
COPY --from=pingcap/binary /src/bin/subreaper /usr/local/bin/subreaper