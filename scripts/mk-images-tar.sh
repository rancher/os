#!/bin/bash
set -ex

suffix=""
[ "$ARCH" == "amd64" ] || suffix="_${ARCH}"

cd $(dirname $0)/..
. scripts/build-common

images="$(build/host_ros c images -i os-config${suffix}.yml)"
for i in ${images}; do
    [ "${FORCE_PULL}" != "1" ] && docker inspect $i >/dev/null 2>&1 || docker pull $i;
done

docker save ${images} > ${BUILD}/images.tar
