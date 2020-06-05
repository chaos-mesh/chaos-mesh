FROM golang:alpine3.10 AS build_base

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add --no-cache gcc g++ make bash git
RUN apk add --update nodejs yarn

ENV GO111MODULE=on
WORKDIR /src
COPY go.mod .
COPY go.sum .

RUN go mod download

FROM build_base AS binary_builder

ARG HTTPS_PROXY
ARG HTTP_PROXY
ARG UI
ARG SWAGGER

COPY . /src
WORKDIR /src
RUN make binary
