FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add tzdata --no-cache

COPY build /web

COPY --from=pingcap/binary /src/bin/chaos-dashboard /usr/local/bin/chaos-dashboard