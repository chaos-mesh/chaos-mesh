FROM alpine:3.10

ARG HTTPS_PROXY
ARG HTTP_PROXY
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add tzdata --no-cache
RUN mkdir -p /usr/local/bin/config/webhook/
COPY --from=pingcap/binary /src/config/webhook/ /usr/local/bin/config/webhook/
COPY --from=pingcap/binary /src/bin/chaos-controller-manager /usr/local/bin/chaos-controller-manager