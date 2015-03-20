#!/bin/bash
set -x -e

MOUNT_POINT=/media/config-2
CONFIG_DEV=$(blkid | grep -- 'LABEL="config-2"' | cut -f1 -d:)

mkdir -p ${MOUNT_POINT}

if [ -e "${CONFIG_DEV}" ]; then
    mount ${CONFIG_DEV} ${MOUNT_POINT}
else
    mount -t 9p -o trans=virtio,version=9p2000.L config-2 ${MOUNT_POINT} 2>/dev/null || true
fi

rancherctl config get cloud_init

cloud-init -save -network=${CLOUD_INIT_NETWORK:-true}
