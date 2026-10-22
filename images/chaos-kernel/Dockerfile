# syntax=docker/dockerfile:1
FROM ubuntu:24.04

ARG HTTPS_PROXY
ARG HTTP_PROXY
ARG MIRROR
ARG TARGETARCH

RUN --mount=type=cache,id=chaos-mesh-ubuntu24.04-apt-${TARGETARCH},target=/var/cache/apt,sharing=locked \
    rm -f /etc/apt/apt.conf.d/docker-clean \
    && if [ -n "$MIRROR" ]; then \
      sed -i -e "s|http://archive.ubuntu.com/ubuntu|$MIRROR|g" \
        -e "s|http://ports.ubuntu.com/ubuntu-ports|$MIRROR|g" /etc/apt/sources.list.d/ubuntu.sources; \
    fi \
    && apt-get -o Acquire::Retries=3 update \
    && DEBIAN_FRONTEND=noninteractive apt-get -o Acquire::Retries=3 install -y --no-install-recommends \
      bison build-essential ca-certificates clang-18 cmake flex git \
      libabsl-dev libclang-18-dev libclang-cpp18-dev libedit-dev libelf-dev \
      libfl-dev libgflags-dev libgrpc++-dev libprotobuf-dev libspdlog-dev \
      libssl-dev llvm-18-dev pkg-config protobuf-compiler protobuf-compiler-grpc \
      python3 python3-dev python3-setuptools zlib1g-dev \
    && rm -rf /var/lib/apt/lists/*

ARG MAKE_JOBS=4
ARG BCC_VERSION=v0.37.0
RUN git clone --depth 1 --branch ${BCC_VERSION} --recurse-submodules=src/cc/libbpf https://github.com/iovisor/bcc.git /bcc
WORKDIR /bcc/build
RUN cmake .. -DCMAKE_INSTALL_PREFIX=/usr -DPYTHON_CMD=python3 \
      -DLLVM_DIR=/usr/lib/llvm-18/lib/cmake/llvm -DENABLE_LLVM_SHARED=ON \
      -DENABLE_TESTS=OFF \
    && cmake --build . --parallel ${MAKE_JOBS} \
    && cmake --install . \
    && ldconfig

ARG BPFKI_COMMIT=8f1e58b18045ace126ff2b232bdd30a21b4fb96e
WORKDIR /bpfki
RUN test -n "${BPFKI_COMMIT}" \
    && git init . \
    && git remote add origin https://github.com/chaos-mesh/bpfki.git \
    && git fetch --depth 1 origin ${BPFKI_COMMIT} \
    && git checkout --detach FETCH_HEAD
RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build --parallel ${MAKE_JOBS} \
    && cmake --install build --prefix /usr/local

WORKDIR /usr/local/bin
