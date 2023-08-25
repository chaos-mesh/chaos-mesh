FROM golang:1.20-alpine3.18

WORKDIR /src

COPY . /src

RUN go build -o test .

FROM alpine:3.18

COPY --from=0 /src/test /bin

ENTRYPOINT ["/bin/test"]
