FROM avcosystems/golang-node AS build_base

ARG HTTPS_PROXY
ARG HTTP_PROXY

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
