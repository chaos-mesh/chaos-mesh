FROM alpine:3.5

RUN apk add tzdata --no-cache
RUN mkdir -p /scripts
ADD hack/wait-fuse.sh /scripts/wait-fuse.sh
