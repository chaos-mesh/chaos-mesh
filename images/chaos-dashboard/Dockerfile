FROM debian:bookworm-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

ARG UI
ARG SWAGGER

RUN echo "deb http://security.debian.org/debian-security bookworm-security main contrib non-free" >> /etc/apt/sources.list && \
    echo "deb http://deb.debian.org/debian bookworm-proposed-updates main contrib non-free" >> /etc/apt/sources.list && \
    apt --allow-releaseinfo-change update && apt upgrade -y && \
    apt install tzdata ca-certificates -y && \
    rm -rf /var/lib/apt/lists/*

ENV http_proxy=
ENV https_proxy=

COPY bin/chaos-dashboard /usr/local/bin/chaos-dashboard
