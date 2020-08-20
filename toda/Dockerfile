# syntax=docker/dockerfile:experimental

FROM debian:buster-slim

ARG HTTPS_PROXY
ARG HTTP_PROXY

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install libfuse-dev pkg-config fuse curl build-essential -y
RUN apt-get install gdb vim -y

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- --default-toolchain nightly-2020-07-01 -y
ENV PATH "/root/.cargo/bin:${PATH}"
ENV RUSTFLAGS "-Z relro-level=full" 

WORKDIR /toda-build

COPY . .

RUN --mount=type=cache,target=/toda-build/target \
    cargo build

RUN --mount=type=cache,target=/toda-build/target \
    cp /toda-build/target/debug/toda /toda