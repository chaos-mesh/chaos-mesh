FROM alpine:3.5

RUN apk add tzdata --no-cache

COPY build /web

ADD bin/chaos-dashboard /usr/local/bin/chaos-dashboard