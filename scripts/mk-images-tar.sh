#!/bin/bash
set -ex

cd $(dirname $0)/..
. scripts/build-common

images="$(build/host_ros c images -i build/os-config.yml)"
for i in ${images}; do
    [ "${FORCE_PULL}" != "1" ] && docker inspect $i >/dev/null 2>&1 || docker pull $i;
done

docker save ${images} > ${BUILD}/images.tar
