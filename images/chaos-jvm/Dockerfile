FROM maven:3.6.3-jdk-8 AS jvmchaos_build

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV http_proxy $HTTP_PROXY
ENV https_proxy $HTTPS_PROXY

RUN apt-get update && apt-get install -y make && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /opt/sandbox

WORKDIR /opt/sandbox

RUN curl -fsSL -o /opt/sandbox/jvmchaos.zip https://github.com/chaosblade-io/chaosblade-exec-jvm/archive/v0.9.0.zip && unzip jvmchaos.zip

WORKDIR /opt/sandbox/chaosblade-exec-jvm-0.9.0

RUN make

FROM alpine:3.12

WORKDIR /bin

COPY --from=jvmchaos_build /opt/sandbox/chaosblade-exec-jvm-0.9.0/build-target/chaosblade-0.9.0/lib/ /bin
