#!/bin/bash

# How to use:
#   make shell-bind
#   cd scripts/tools/
#   wget https://link/rancheros-xxx.img
#   wget http://link/custom.crt
#
#   ./flush_crt_nbd.sh --img rancheros-openstack.img --cert custom.crt
#   # or
#   ./flush_crt_nbd.sh --initrd initrd-xxxx --cert custom.crt
#
#   exit
#   ls ./build/
#

set -ex

BASE_DIR=/tmp
ORIGIN_DIR=/tmp/origin
WORK_DIR=/tmp/work

mkdir -p ${ORIGIN_DIR} ${WORK_DIR} ${DAPPER_SOURCE}/build

while [ "$#" -gt 0 ]; do
    case $1 in
        --initrd)
            shift 1
            INITRD_FILE=$(readlink -f $1)
            ;;
        --img)
            shift 1
            IMG_FILE=$(readlink -f $1)
            ;;
        --cert)
            shift 1
            CERT_FILE=$(readlink -f $1)
            ;;
        *)
            break
            ;;
    esac
    shift 1
done

function rebuild_initrd() {
    local initrd_name=$1
    local output_dir=$2

    # update and rebuild the initrd
    pushd ${WORK_DIR}
    mv initrd-* ${initrd_name}.gz
    gzip -d ${initrd_name}.gz
    cpio -i -F ${initrd_name}
    rm -f ${initrd_name}
    cat ${CERT_FILE} >> ${WORK_DIR}/usr/etc/ssl/certs/ca-certificates.crt
    find | cpio -H newc -o | gzip -9 > ${output_dir}/${initrd_name}
    popd
}


if [ ! -z ${INITRD_FILE} ]; then
    cp ${INITRD_FILE} ${WORK_DIR}/
    rebuild_initrd $(basename ${INITRD_FILE}) ${DAPPER_SOURCE}/build/
    exit 0
fi

# copy the iso content
cp -a ${IMG_FILE} ${IMG_FILE}_new
qemu-nbd -c /dev/nbd0 --partition=1 ${IMG_FILE}_new
mount /dev/nbd0 ${ORIGIN_DIR}

# copy the initrd file
INITRD_NAME=$(basename ${ORIGIN_DIR}/boot/initrd-*)
cp ${ORIGIN_DIR}/boot/initrd-* ${WORK_DIR}/

rebuild_initrd ${INITRD_NAME} ${ORIGIN_DIR}/boot

# copy out
umount ${ORIGIN_DIR}
qemu-nbd -d /dev/nbd0
mv ${IMG_FILE}_new ${DAPPER_SOURCE}/build/$(basename ${IMG_FILE})

# cleanup
rm -rf ${WORK_DIR}/
