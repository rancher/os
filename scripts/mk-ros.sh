#!/bin/bash
set -ex

ros="$1"

ARCH=${ARCH:?"ARCH not set"}
VERSION=${VERSION:?"VERSION not set"}

case "$ARCH" in
	"arm")
		GCC_PACKAGE="arm-linux-gnueabihf";;
	"arm64")
		GCC_PACKAGE="aarch64-linux-gnu";;
	"ppc64le")
		GCC_PACKAGE="powerpc64le-linux-gnu";;
esac

cd $(dirname $0)/..

STRIP_BIN=$(which strip)
if [[ "${ARCH}" != "amd64" ]]; then
  export GOARM=7
  export CC=/usr/bin/${GCC_PACKAGE}-gcc
  export CGO_ENABLED=1
  STRIP_BIN=/usr/${GCC_PACKAGE}/bin/strip
fi
GOARCH=${ARCH} go build -tags netgo -installsuffix netgo -ldflags "-X github.com/rancher/os/config.VERSION=${VERSION} -linkmode external -extldflags -static" -o ${ros}
${STRIP_BIN} --strip-all ${ros}
