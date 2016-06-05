#!/bin/bash
set -ex

TARGET=$(pwd)/${1}

SUFFIX=${SUFFIX:-""}
HOST_SUFFIX=${HOST_SUFFIX:-""}
DFS_IMAGE=${DFS_IMAGE:?"DFS_IMAGE not set"}
IS_ROOTFS=${IS_ROOTFS:-0}

cd $(dirname $0)/..
. scripts/build-common

INITRD_DIR=${BUILD}/initrd

rm -rf ${INITRD_DIR}/{usr,init}
mkdir -p ${INITRD_DIR}/usr/{bin,share/ros}
mkdir -p ${INITRD_DIR}/var/lib/system-docker
mkdir -p ${INITRD_DIR}/usr/etc/selinux/ros/{policy,contexts}

if [ "$IS_ROOTFS" == "0" ]; then
  cp -rf ${BUILD}/kernel/lib ${INITRD_DIR}/usr/
fi
cp assets/docker           ${INITRD_DIR}/usr/bin/docker
if [ "$IS_ROOTFS" == "0" ]; then
  cp ${BUILD}/images.tar     ${INITRD_DIR}/usr/share/ros/
fi
cp build/os-config.yml     ${INITRD_DIR}/usr/share/ros/
cp bin/ros                 ${INITRD_DIR}/usr/bin/
ln -s usr/bin/ros          ${INITRD_DIR}/init
ln -s bin                  ${INITRD_DIR}/usr/sbin
ln -s usr/sbin             ${INITRD_DIR}/sbin

cp assets/selinux/config            ${INITRD_DIR}/usr/etc/selinux/
cp assets/selinux/policy.29         ${INITRD_DIR}/usr/etc/selinux/ros/policy/
cp assets/selinux/seusers           ${INITRD_DIR}/usr/etc/selinux/ros/
cp assets/selinux/lxc_contexts      ${INITRD_DIR}/usr/etc/selinux/ros/contexts/
cp assets/selinux/failsafe_context  ${INITRD_DIR}/usr/etc/selinux/ros/contexts/

if [ "$ARCH" == "amd64" ]; then
  KERNEL_RELEASE=$(tar xvf assets/modules.tar.gz -C ${INITRD_DIR} | cut -f4 -d/ | cut -f1 -d ' ')
  depmod -a -b ${INITRD_DIR}/usr $KERNEL_RELEASE
fi

if [ -e assets/extra.tar.gz ]; then
  KERNEL_RELEASE=$(tar xvf assets/extra.tar.gz -C ${INITRD_DIR}/usr | cut -f4 -d/ | cut -f1 -d ' ')
  rm -rf ${INITRD_DIR}/usr/lib/firmware/.git
  depmod -a -b ${INITRD_DIR}/usr $KERNEL_RELEASE
fi

DFS_ARCH=$(docker create ${DFS_IMAGE}${SUFFIX})
trap "docker rm -fv ${DFS_ARCH}" EXIT

docker export ${DFS_ARCH} | tar xvf - -C ${INITRD_DIR} --exclude=usr/bin/dockerlaunch \
                                                       --exclude=usr/bin/docker       \
                                                       --exclude=usr/share/git-core   \
                                                       --exclude=usr/bin/git          \
                                                       --exclude=usr/bin/ssh          \
                                                       --exclude=usr/libexec/git-core \
                                                       usr

if [ "$IS_ROOTFS" == "1" ]; then
  DFS=$(docker run -d --privileged -v /lib/modules/$(uname -r):/lib/modules/$(uname -r) ${DFS_IMAGE}${HOST_SUFFIX})
  trap "docker rm -fv ${DFS_ARCH} ${DFS}" EXIT
  docker exec -i ${DFS} docker load < ${BUILD}/images.tar
  docker stop ${DFS}
  docker run --rm --volumes-from=${DFS} rancher/os-dapper-base tar -c -C /var/lib/docker ./image | tar -x -C ${INITRD_DIR}/var/lib/system-docker
  docker run --rm --volumes-from=${DFS} rancher/os-dapper-base tar -c -C /var/lib/docker ./overlay | tar -x -C ${INITRD_DIR}/var/lib/system-docker

  cd ${INITRD_DIR}

  tar -czf ${TARGET} .
else
  COMPRESS=lzma
  [ "$DEV_BUILD" == "1" ] && COMPRESS="gzip -1"

  cd ${INITRD_DIR}

  find | cpio -H newc -o | ${COMPRESS} > ${TARGET}
fi
