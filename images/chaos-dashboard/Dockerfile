FROM alpine:3.23

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy=$HTTP_PROXY
ENV https_proxy=$HTTPS_PROXY

ARG UI
ARG SWAGGER

RUN apk upgrade --no-cache && apk add --no-cache tzdata ca-certificates

ENV http_proxy=
ENV https_proxy=

COPY bin/chaos-dashboard /usr/local/bin/chaos-dashboard
