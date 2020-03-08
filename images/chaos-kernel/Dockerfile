FROM ubuntu:bionic

ARG HTTPS_PROXY
ARG HTTP_PROXY

RUN apt update && apt install -y ca-certificates gnupg2

RUN echo "deb https://mirrors.aliyun.com/ubuntu/ bionic main restricted universe multiverse\n"\
"deb https://mirrors.aliyun.com/ubuntu/ bionic-security main restricted universe multiverse\n"\
"deb https://mirrors.aliyun.com/ubuntu/ bionic-updates main restricted universe multiverse\n"\
"deb https://mirrors.aliyun.com/ubuntu/ bionic-proposed main restricted universe multiverse\n"\
"deb https://mirrors.aliyun.com/ubuntu/ bionic-backports main restricted universe multiverse\n"\
"deb-src https://mirrors.aliyun.com/ubuntu/ bionic main restricted universe multiverse\n"\
"deb-src https://mirrors.aliyun.com/ubuntu/ bionic-security main restricted universe multiverse\n"\
"deb-src https://mirrors.aliyun.com/ubuntu/ bionic-updates main restricted universe multiverse\n"\
"deb-src https://mirrors.aliyun.com/ubuntu/ bionic-proposed main restricted universe multiverse\n"\
"deb-src https://mirrors.aliyun.com/ubuntu/ bionic-backports main restricted universe multiverse\n"\
"deb https://apt.kitware.com/ubuntu/ bionic main" > /etc/apt/sources.list

RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 8611B7A28669BB93

RUN apt update &&  apt install -y gcc-8 g++-8 bison build-essential \
    flex git libedit-dev libllvm6.0 llvm-6.0-dev libclang-6.0-dev python python-pip \
    zlib1g-dev libelf-dev libssl-dev && pip install cmake==3.13.3

RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-7 70
RUN update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-7 70
RUN update-alternatives --install /usr/bin/gcov gcov /usr/bin/gcov-7 70
RUN update-alternatives --install /usr/bin/gcc gcc /usr/bin/gcc-8 80
RUN update-alternatives --install /usr/bin/g++ g++ /usr/bin/g++-8 80
RUN update-alternatives --install /usr/bin/gcov gcov /usr/bin/gcov-8 80

RUN git clone https://github.com/iovisor/bcc.git
WORKDIR /bcc/build
RUN cmake .. -DCMAKE_INSTALL_PREFIX=/usr && make -j && make install

WORKDIR /

RUN git clone https://github.com/chaos-mesh/bpfki
WORKDIR /bpfki/build
RUN cmake .. && make -j && mv bin/bpfki /usr/local/bin/ && mv examples/fail* /usr/local/bin/

WORKDIR /usr/local/bin
