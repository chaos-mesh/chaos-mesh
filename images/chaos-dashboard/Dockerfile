FROM mhart/alpine-node:13.11.0 as node

RUN mkdir /src
WORKDIR /src
COPY . .
RUN yarn install; yarn build

FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apk add tzdata --no-cache

COPY --from=node /src/build /web

COPY --from=pingcap/binary /src/bin/chaos-dashboard /usr/local/bin/chaos-dashboard