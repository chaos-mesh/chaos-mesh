# syntax=docker/dockerfile:experimental

FROM golang:1.25-alpine AS build-env
RUN go install github.com/go-delve/delve/cmd/dlv@v1.21.0

FROM alpine:3.23

RUN apk add --no-cache procps

COPY --from=build-env /go/bin/dlv /
CMD ["sh", "-c", "/dlv attach $(ps axf | grep $CMD_NAME | grep -v grep | awk '{print $1}') --listen=:8000 --headless=true --api-version=2 --accept-multiclient --continue"]
