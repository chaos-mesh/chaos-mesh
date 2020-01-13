FROM alpine:3.5

RUN apk add tzdata --no-cache
RUN mkdir -p /scripts
ADD scripts/wait-fuse.sh /scripts/wait-fuse.sh
ADD scripts/init.sh /scripts/init.sh
