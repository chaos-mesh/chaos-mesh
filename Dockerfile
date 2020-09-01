# syntax=docker/dockerfile:experimental

FROM pingcap/chaos-build-base AS go_build

RUN curl https://dl.google.com/go/go1.14.6.linux-amd64.tar.gz | tar -xz -C /usr/local
ENV PATH "/usr/local/go/bin:${PATH}"
ENV GO111MODULE=on

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN if [[ -n "$HTTP_PROXY" ]]; then yarn config set proxy $HTTP_PROXY; fi

WORKDIR /src

COPY . /src

ARG UI
ARG SWAGGER
ARG LDFLAGS

RUN --mount=type=cache,target=/root/go/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/src/ui/node_modules \
    IMG_LDFLAGS=$LDFLAGS make binary

FROM pingcap/chaos-build-base AS rust_build

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- --default-toolchain nightly-2020-07-01 -y
ENV PATH "/root/.cargo/bin:${PATH}"

COPY ./toda /toda-build

WORKDIR /toda-build

RUN if [ -n "$HTTP_PROXY" ]; then echo "[http]\n\
proxy = \"${HTTP_PROXY}\"\n\
"\
> /root/.cargo/config ; fi

ARG CRATES_MIRROR

RUN if [ -n "$CRATES_MIRROR" ]; then echo "\n\
[source.crates-io]\n\
replace-with = 'mirror'\n\
[source.mirror]\n\
registry = \"$CRATES_MIRROR\"\n\
"> /root/.cargo/config ; fi

ENV CARGO_LOG trace
ENV CARGO_HTTP_DEBUG true
ENV CARGO_HTTP_MULTIPLEXING false
ENV RUSTFLAGS "-Z relro-level=full"
RUN --mount=type=cache,target=/toda-build/target \
    --mount=type=cache,target=/root/.cargo/registry \
    cargo build --release

RUN --mount=type=cache,target=/toda-build/target \
    cp /toda-build/target/release/toda /toda

FROM alpine:3.12

COPY ./scripts /scripts
COPY --from=go_build /src/bin /bin
COPY --from=rust_build /toda /bin/toda