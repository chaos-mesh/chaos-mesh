FROM alpine:3.12

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV TZ /etc/localtime
RUN apk add --no-cache tzdata fuse

COPY --from=pingcap/chaos-binary /bin/chaosfs /usr/local/bin/chaosfs
