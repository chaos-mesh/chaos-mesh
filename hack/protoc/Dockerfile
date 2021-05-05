FROM golang:1.14.4-alpine3.12

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV GO111MODULE on
ENV PROTOC_VERSION 3.12.3

RUN apk add --no-cache wget unzip protoc protobuf-dev make bash

RUN go get github.com/golang/protobuf/protoc-gen-go@v1.4.2

RUN mkdir /.cache
RUN chmod -R 777 /.cache
