FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add --no-cache tzdata iptables ipset stress-ng iproute2

COPY --from=pingcap/binary /src/bin/chaos-daemon /usr/local/bin/chaos-daemon