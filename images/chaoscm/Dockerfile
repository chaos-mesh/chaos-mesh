FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV TZ /etc/localtime
RUN apk add --no-cache tzdata

COPY --from=pingcap/binary /src/bin/chaoscm /usr/local/bin/chaoscm
