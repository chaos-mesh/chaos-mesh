FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV TZ /etc/localtime
RUN apk add --no-cache tzdata fuse

COPY --from=pingcap/binary /src/bin/chaosfs /usr/local/bin/chaosfs
