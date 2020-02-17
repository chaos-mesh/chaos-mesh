FROM golang:alpine3.10 AS build_base

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add --no-cache gcc g++ make bash git

ENV GO111MODULE=on
RUN mkdir /src
WORKDIR /src
COPY go.mod .
COPY go.sum .

RUN go mod download
RUN go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4
RUN go get golang.org/x/tools/cmd/goimports

FROM build_base AS binary_builder

ARG HTTPS_PROXY
ARG HTTP_PROXY

COPY . /src
WORKDIR /src
RUN make binary