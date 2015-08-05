#!/bin/bash
set -ex

cd $(dirname $0)/..
. scripts/build-common

mv ${BUILD}/kernel/lib ${INITRD_DIR}
mv assets/docker       ${INITRD_DIR}
cp os-config.yml       ${INITRD_DIR}
cp bin/rancheros       ${INITRD_DIR}/init
cd ${INITRD_DIR} && find | cpio -H newc -o | lzma -c > ${DIST}/artifacts/initrd
