# syntax=docker/dockerfile:experimental
FROM golang:1.14.4-alpine3.12 AS build_base

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

RUN if [[ -n "$HTTP_PROXY" ]]; then yarn config set proxy $HTTP_PROXY; fi

COPY . /src
WORKDIR /src
RUN --mount=type=cache,target=/root/.cache/go-build \
    make binary
