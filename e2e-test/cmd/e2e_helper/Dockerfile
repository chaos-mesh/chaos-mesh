FROM golang:1.23-alpine3.21

WORKDIR /src

COPY . /src

RUN go build -o test .

FROM alpine:3.21

COPY --from=0 /src/test /bin

ENTRYPOINT ["/bin/test"]
