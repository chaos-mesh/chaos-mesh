FROM golang:alpine3.10 AS build

RUN apk update
RUN apk add gcc g++ make bash
ADD src /src
WORKDIR /src
ENV GOPATH /go
RUN make chaosdaemon

FROM alpine:3.10

RUN apk update
RUN apk add tzdata
RUN apk add iptables
RUN apk add ipset
COPY --from=build /src/bin/chaos-daemon /usr/local/bin/chaos-daemon