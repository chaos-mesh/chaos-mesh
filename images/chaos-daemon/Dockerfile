FROM golang:alpine3.10 AS build

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk update
RUN apk add gcc g++ make bash
ADD src /src
WORKDIR /src
ENV GOPATH /go
ENV HTTP_PROXY http://192.168.50.173:8000
ENV HTTPS_PROXY http://192.168.50.173:8000
RUN make chaosdaemon

FROM alpine:3.10

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk update
RUN apk add tzdata
RUN apk add iptables
RUN apk add ipset
COPY --from=build /src/bin/chaos-daemon /usr/local/bin/chaos-daemon