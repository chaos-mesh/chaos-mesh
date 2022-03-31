FROM golang:1.18-alpine3.15

WORKDIR /src

COPY main.go /src
COPY go.mod /src
COPY go.sum /src

RUN go build -o test main.go

FROM alpine:3.15

COPY --from=0 /src/test /bin

ENTRYPOINT ["/bin/test"]
