FROM alpine:3

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk upgrade --no-cache && apk add --no-cache tzdata

COPY bin/chaos-controller-manager /usr/local/bin/chaos-controller-manager
