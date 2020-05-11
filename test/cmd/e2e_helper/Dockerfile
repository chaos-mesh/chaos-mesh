FROM golang:alpine3.10

WORKDIR /src

COPY main.go /src

RUN go build -o test main.go

FROM alpine:3.10

COPY --from=0 /src/test /bin

ENTRYPOINT ["/bin/test"]
