FROM alpine:3.12

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add tzdata --no-cache

COPY --from=pingcap/chaos-binary /bin/chaos-controller-manager /usr/local/bin/chaos-controller-manager