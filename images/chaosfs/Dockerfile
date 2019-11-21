FROM alpine:3.5

ENV TZ /etc/localtime
RUN apk add --no-cache tzdata fuse
ADD bin/chaosfs /usr/local/bin/chaosfs
