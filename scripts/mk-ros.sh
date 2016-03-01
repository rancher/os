#!/bin/bash
set -ex

ros="$1"

ARCH=${ARCH:?"ARCH not set"}
VERSION=${VERSION:?"VERSION not set"}

cd $(dirname $0)/..

[ "${ARCH}" == "arm" ] && export GOARM=6
GOARCH=${ARCH} go build -tags netgo -installsuffix netgo -ldflags "-X github.com/rancher/os/config.VERSION=${VERSION} -linkmode external -extldflags -static" -o ${ros}
strip --strip-all ${ros}
