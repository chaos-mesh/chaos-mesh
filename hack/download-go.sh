#!/bin/bash

ARCH="$(uname -m)"
case $ARCH in
"x86_64")
    ARCH=amd64
    ;;
"aarch64")
    ARCH=arm64
    ;;
"armv6")
    ARCH=armv6l
    ;;
"armv8")
    ARCH=arm64
    ;;
.*386.*)
    ARCH=386
    ;;
esac

echo "Downloading for $ARCH"
curl -k https://dl.google.com/go/go1.14.6.linux-$ARCH.tar.gz | tar -xz -C /usr/local
