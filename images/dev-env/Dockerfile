# syntax=docker/dockerfile:experimental

FROM debian:buster-slim

ENV DEBIAN_FRONTEND noninteractive

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install protobuf-compiler git build-essential curl -y

ARG TARGET_PLATFORM=amd64
RUN curl https://dl.google.com/go/go1.15.11.linux-$TARGET_PLATFORM.tar.gz | tar -xz -C /usr/local
ENV GO111MODULE=on
ENV GOPATH /go
ENV PATH "/usr/local/go/bin:${PATH}:/tmp/go/bin:/go/bin"

RUN go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1
RUN go get github.com/mgechev/revive@v1.0.2-0.20200225072153-6219ca02fffb
RUN go get github.com/pingcap/failpoint/failpoint-ctl@v0.0.0-20200210140405-f8f9fb234798
RUN go get golang.org/x/tools/cmd/goimports@v0.1.4
RUN go get github.com/securego/gosec/cmd/gosec@v0.0.0-20200401082031-e946c8c39989
RUN go get github.com/99designs/gqlgen@v0.13.0
RUN go get github.com/golang/protobuf/protoc-gen-go@v1.4.2
RUN go get github.com/axw/gocov/gocov@v1.0.0

RUN curl -L https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.5.4/kustomize_v3.5.4_$(go env GOOS)_$(go env GOARCH).tar.gz | tar -xz -C /usr/local/bin/
RUN curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v2.3.2/kubebuilder_2.3.2_$(go env GOOS)_$(go env GOARCH).tar.gz | tar -xz -C /usr/local/
RUN curl -L https://get.helm.sh/helm-v3.6.3-$(go env GOOS)-$(go env GOARCH).tar.gz | tar -xz -C /usr/local/bin
RUN mv /usr/local/kubebuilder_2.3.2_$(go env GOOS)_$(go env GOARCH) /usr/local/kubebuilder
RUN mv /usr/local/bin/$(go env GOOS)-$(go env GOARCH)/helm /usr/local/bin/helm
RUN rm -rf /usr/local/bin/$(go env GOOS)-$(go env GOARCH)

ENV GOCACHE /tmp/go-build
ENV GOPATH /tmp/go
