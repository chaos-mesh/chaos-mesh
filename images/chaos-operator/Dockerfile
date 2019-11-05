FROM alpine:3.5

RUN apk add tzdata --no-cache
ADD bin/chaos-controller-manager /usr/local/bin/chaos-controller-manager
ADD bin/chaos-admission-controller /usr/local/bin/chaos-admission-controller