FROM alpine:3.5

RUN apk add tzdata --no-cache
RUN apk add iptables --no-cache
RUN apk add ipset --no-cache
ADD bin/chaos-daemon /usr/local/bin/chaos-daemon