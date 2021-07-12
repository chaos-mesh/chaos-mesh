# syntax=docker/dockerfile:experimental

FROM golang:1.15.11 AS build-env

ENV GO111MODULE on

WORKDIR /
RUN wget https://github.com/go-delve/delve/archive/refs/tags/v1.6.1.tar.gz
RUN tar xvf v1.6.1.tar.gz
WORKDIR /delve-1.6.1
RUN go install github.com/go-delve/delve/cmd/dlv

FROM debian:buster

RUN apt-get update
RUN apt-get install procps -y

COPY --from=build-env /go/bin/dlv /
CMD ["bash", "-c", "/dlv attach $(ps axf | grep $CMD_NAME | grep -v grep | awk '{print $1}') --listen=:8000 --headless=true --api-version=2 --accept-multiclient --continue"]