# syntax=docker/dockerfile:1
# Copyright 2026 Chaos Mesh Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

FROM scratch AS base
ADD base.tar.gz /
COPY inputs.lock.json /opt/inputs.lock.json
RUN rm -f /etc/apt/sources.list /etc/apt/sources.list.d/* /etc/apt/apt.conf.d/docker-clean && \
    printf '%s\n' \
      'deb [signed-by=/usr/share/keyrings/ubuntu-archive-keyring.gpg] http://archive.ubuntu.com/ubuntu noble main universe' \
      'deb [signed-by=/usr/share/keyrings/ubuntu-archive-keyring.gpg] http://archive.ubuntu.com/ubuntu noble-updates main universe' \
      'deb [signed-by=/usr/share/keyrings/ubuntu-archive-keyring.gpg] http://security.ubuntu.com/ubuntu noble-security main universe' \
      > /etc/apt/sources.list && \
    printf '#!/bin/sh\nexit 101\n' > /usr/sbin/policy-rc.d && chmod 755 /usr/sbin/policy-rc.d

FROM base AS fixture
RUN --mount=type=cache,id=chaos-mesh-ubuntu24.04-apt-amd64,target=/var/cache/apt,sharing=locked \
    apt-get -o Acquire::Retries=3 update && \
    DEBIAN_FRONTEND=noninteractive apt-get -o Acquire::Retries=3 install -y --no-install-recommends g++ libc6-dev
COPY lifecycle/fixture.cc /src/fixture.cc
COPY filter_layout.h /src/filter_layout.h
ARG FILTER_HEADER_SHA256
RUN mkdir /out && \
    g++ --version > /out/fixture-toolchain.txt && \
    g++ -dumpmachine >> /out/fixture-toolchain.txt && \
    ld --version >> /out/fixture-toolchain.txt && \
    dpkg-query -W -f='${Package} ${Version} ${Architecture}\n' g++ gcc libc6-dev libstdc++-13-dev \
      >> /out/fixture-toolchain.txt && \
    g++ -std=c++17 -O2 -static -Wall -Wextra -Werror -I /src \
      -DCM_FILTER_HEADER_SHA256=\"${FILTER_HEADER_SHA256}\" /src/fixture.cc -o /out/cm-lifecycle-fixture

FROM base AS node
COPY package-seal.py /opt/package-seal.py
RUN --mount=type=cache,id=chaos-mesh-ubuntu24.04-apt-amd64,target=/var/cache/apt,sharing=locked \
    mkdir -p /var/lib/cm-lifecycle-seal && \
    dpkg-query -W -f='${Package}\t${Version}\t${Architecture}\t${db:Status-Status}\n' \
      > /var/lib/cm-lifecycle-seal/base-manifest.tsv && \
    apt-get -o Acquire::Retries=3 update && \
    DEBIAN_FRONTEND=noninteractive apt-get -o Acquire::Retries=3 \
      -o Binary::apt::APT::Keep-Downloaded-Packages=true -o APT::Keep-Downloaded-Packages=true \
      install --no-install-recommends -y systemd systemd-sysv udev dbus iproute2 iptables nftables \
      util-linux mount procps python3 python3-yaml coreutils ca-certificates && \
    python3 /opt/package-seal.py && \
    rm -rf /var/lib/apt/lists/*
COPY tools/ /
COPY lifecycle/ /opt/cm-lifecycle/
COPY root-overlay/ /
RUN chmod 755 /opt/cm-lifecycle/init.sh && \
    update-alternatives --set iptables /usr/sbin/iptables-nft && \
    update-alternatives --set ip6tables /usr/sbin/ip6tables-nft && \
    python3 /opt/cm-lifecycle/node.py audit-rootfs && \
    rm /opt/package-seal.py /usr/sbin/policy-rc.d
ENTRYPOINT ["/opt/cm-lifecycle/init.sh"]
