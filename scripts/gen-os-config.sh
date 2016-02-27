#!/bin/bash
set -ex

cd $(dirname $0)/..

set -a
. build.conf

SUFFIX=""
[ "${ARCH}" == "amd64" ] || SUFFIX="_${ARCH}"
set +a

build/host_ros c generate < os-config.tpl.yml > $1
