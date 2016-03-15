#!/bin/bash
set -ex

ros="$1"

ARCH=${ARCH:?"ARCH not set"}
VERSION=${VERSION:?"VERSION not set"}

cd $(dirname $0)/..

strip_bin=$(which strip)
[ "${ARCH}" == "arm" ] && export GOARM=6
if [ "${TOOLCHAIN}" != "" ]; then
  export CC=/usr/bin/${TOOLCHAIN}-gcc
  export CGO_ENABLED=1
  strip_bin=/usr/bin/${TOOLCHAIN}-strip
fi
GOARCH=${ARCH} go build -tags netgo -installsuffix netgo -ldflags "-X github.com/rancher/os/config.VERSION=${VERSION} -linkmode external -extldflags -static" -o ${ros}
${strip_bin} --strip-all ${ros}
