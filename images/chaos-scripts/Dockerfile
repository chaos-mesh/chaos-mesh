FROM alpine:3.12

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add tzdata --no-cache
RUN mkdir -p /scripts

COPY --from=pingcap/chaos-binary /scripts/wait-fuse.sh /scripts/wait-fuse.sh
COPY --from=pingcap/chaos-binary /scripts/init.sh /scripts/init.sh
