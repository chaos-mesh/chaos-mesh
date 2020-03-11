FROM alpine:3.10

RUN apk add --no-cache tzdata stress-ng

COPY --from=pingcap/binary /src/bin/chaos-stress /usr/local/bin/chaos-stress