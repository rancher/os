#!/bin/bash
set -ex

ros="$1"

ARCH=${ARCH:?"ARCH not set"}
VERSION=${VERSION:?"VERSION not set"}

cd $(dirname $0)/..

strip_bin=$(which strip)
if [ "${ARCH}" == "arm" ]; then
  export GOARM=6
  export CC=/usr/bin/arm-linux-gnueabihf-gcc
  export CGO_ENABLED=1
  strip_bin=/usr/arm-linux-gnueabihf/bin/strip
fi
GOARCH=${ARCH} go build -tags netgo -installsuffix netgo -ldflags "-X github.com/rancher/os/config.VERSION=${VERSION} -linkmode external -extldflags -static" -o ${ros}
${strip_bin} --strip-all ${ros}
