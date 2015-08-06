#!/bin/bash
set -ex

cd $(dirname $0)/..
. scripts/build-common

rm -rf ${INITRD_DIR}/{usr,init}
mkdir -p ${INITRD_DIR}/usr/{bin,share/ros}

cp -rf ${BUILD}/kernel/lib ${INITRD_DIR}/usr
cp assets/docker           ${INITRD_DIR}/usr/bin/docker
cp ${BUILD}/images.tar     ${INITRD_DIR}/usr/share/ros
cp os-config.yml           ${INITRD_DIR}/usr/share/ros/
cp bin/rancheros           ${INITRD_DIR}/usr/bin/ros
ln -s usr/bin/ros          ${INITRD_DIR}/init
ln -s bin                  ${INITRD_DIR}/usr/sbin

docker export $(docker create rancher/docker:1.8.0-rc2) | tar xvf - -C ${INITRD_DIR} --exclude=usr/bin/dockerlaunch \
                                                                                     --exclude=usr/bin/docker       \
                                                                                     --exclude=usr/share/git-core   \
                                                                                     --exclude=usr/bin/git          \
                                                                                     --exclude=usr/bin/ssh          \
                                                                                     --exclude=usr/libexec/git-core \
                                                                                     usr

cd ${INITRD_DIR} && find | cpio -H newc -o | lzma -c > ${DIST}/artifacts/initrd
