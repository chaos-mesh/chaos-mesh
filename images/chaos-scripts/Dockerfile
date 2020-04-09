FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add tzdata --no-cache
RUN mkdir -p /scripts

COPY --from=pingcap/binary /src/scripts/wait-fuse.sh /scripts/wait-fuse.sh
COPY --from=pingcap/binary /src/scripts/init.sh /scripts/init.sh
