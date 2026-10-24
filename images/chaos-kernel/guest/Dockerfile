# syntax=docker/dockerfile:1
ARG KERNEL_IMAGE
FROM ${KERNEL_IMAGE}

COPY manifest.py /opt/image-manifest.py
COPY runtime.json /opt/image-runtime.json
RUN python3 /opt/image-manifest.py check /opt/image-runtime.json /
RUN --mount=type=cache,id=chaos-mesh-ubuntu24.04-apt-amd64,target=/var/cache/apt,sharing=locked \
    apt-get -o Acquire::Retries=3 update && \
    DEBIAN_FRONTEND=noninteractive apt-get -o Acquire::Retries=3 install --no-install-recommends -y \
      sudo iproute2 util-linux busybox-static && \
    rm -rf /var/lib/apt/lists/* && \
    python3 /opt/image-manifest.py check /opt/image-runtime.json /

RUN useradd --uid 1001 --create-home runner && \
    printf 'runner ALL=(ALL) NOPASSWD: ALL\n' > /etc/sudoers.d/runner && \
    chmod 440 /etc/sudoers.d/runner && \
    mkdir -p /home/runner/work/_temp /kernel-headers /boot && \
    chown -R runner:runner /home/runner/work
COPY candidate /opt/bpfki
COPY source-control /opt/source-control
COPY source-control.json /home/runner/work/_temp/bpfki-source-control.json
RUN cmake -S /opt/source-control -B /tmp/control-build -DCMAKE_BUILD_TYPE=Release && \
    cmake --build /tmp/control-build --parallel 2 && \
    cmake --install /tmp/control-build --prefix /opt/control && \
    cmake -S /opt/bpfki -B /tmp/driver-build -DCMAKE_BUILD_TYPE=Release -DBPFKI_BUILD_KERNEL_TESTS=ON && \
    cmake --build /tmp/driver-build --target kernel-tests --parallel 2 && \
    cp /tmp/driver-build/kernel-tests /opt/kernel-tests && \
    cp /tmp/control-build/CMakeCache.txt /home/runner/work/_temp/control-CMakeCache.txt && \
    cp /tmp/driver-build/CMakeCache.txt /home/runner/work/_temp/driver-CMakeCache.txt && \
    ln -s /usr/local /opt/candidate && \
    rm -rf /tmp/control-build /tmp/driver-build /opt/source-control && \
    python3 /opt/image-manifest.py check /opt/image-runtime.json / && \
    python3 /opt/image-manifest.py check-build-paths /opt/image-runtime.json && \
    rm -rf /bcc /bpfki/build && \
    python3 /opt/image-manifest.py check /opt/image-runtime.json / && \
    chmod 755 /opt/bpfki/tests/kernel/vm/init.sh
ENTRYPOINT ["/opt/bpfki/tests/kernel/vm/init.sh"]
