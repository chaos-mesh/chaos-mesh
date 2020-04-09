FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add tzdata --no-cache

COPY --from=pingcap/binary /src/bin/chaos-controller-manager /usr/local/bin/chaos-controller-manager