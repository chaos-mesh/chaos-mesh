FROM golang:1.25.8-alpine3.23

WORKDIR /src

COPY . /src

RUN go build -o test .

FROM alpine:3.23

COPY --from=0 /src/test /bin

ENTRYPOINT ["/bin/test"]
