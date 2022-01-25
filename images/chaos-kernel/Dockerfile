FROM ubuntu:bionic

ARG HTTPS_PROXY
ARG HTTP_PROXY
ARG MAKE_JOBS=4
ARG MIRROR=http://archive.ubuntu.com/ubuntu

RUN apt-get update && apt-get install -y ca-certificates gnupg2 wget

RUN if [ ! -z "$MIRROR" ]; then sed -i "s|http://archive.ubuntu.com/ubuntu|$MIRROR|g" /etc/apt/sources.list; fi

RUN echo "" >> /etc/apt/sources.list
RUN echo "deb https://apt.kitware.com/ubuntu/ bionic main" >> /etc/apt/sources.list

RUN wget -O - https://apt.kitware.com/keys/kitware-archive-latest.asc 2>/dev/null | apt-key add -

RUN ulimit -n 1024 && apt-get update &&  apt-get install -y gcc-8 g++-8 bison build-essential \
    flex git libedit-dev libllvm6.0 llvm-6.0-dev libclang-6.0-dev python python-pip \
    zlib1g-dev libelf-dev libssl-dev
RUN apt install -y cmake

RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-7 70
RUN update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-7 70
RUN update-alternatives --install /usr/bin/gcov gcov /usr/bin/gcov-7 70
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80
RUN update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-8 80
RUN update-alternatives --install /usr/bin/gcov gcov /usr/bin/gcov-8 80

RUN git clone --depth 1 --branch v0.23.0 https://github.com/iovisor/bcc.git
WORKDIR /bcc/build
RUN cmake .. -DCMAKE_INSTALL_PREFIX=/usr && make -j${MAKE_JOBS} && make install

WORKDIR /

RUN git clone https://github.com/chaos-mesh/bpfki
WORKDIR /bpfki/build
RUN cmake .. && make -j${MAKE_JOBS} && mv bin/bpfki /usr/local/bin/ && mv examples/fail* /usr/local/bin/

WORKDIR /usr/local/bin
