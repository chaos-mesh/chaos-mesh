FROM golang:alpine3.10

WORKDIR /src

COPY timer/main.go /src

RUN go build -o timer main.go

FROM alpine:3.10

COPY --from=0 /src/timer /bin

CMD ["/bin/timer"]

