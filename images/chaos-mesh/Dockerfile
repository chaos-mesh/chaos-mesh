FROM alpine:3.23

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk upgrade --no-cache && apk add tzdata --no-cache

COPY bin/chaos-controller-manager /usr/local/bin/chaos-controller-manager
