FROM golang:1.24.4-alpine3.22

WORKDIR /src

COPY . /src

RUN go build -o test .

FROM alpine:3.22

COPY --from=0 /src/test /bin

ENTRYPOINT ["/bin/test"]
