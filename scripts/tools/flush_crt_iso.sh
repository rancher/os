#!/bin/bash

# How to use:
#   make shell-bind
#   cd scripts/tools/
#   wget https://link/rancheros-xxx.iso
#   wget http://link/custom.crt
#
#   ./flush_crt_iso.sh --iso rancheros-vmware-autoformat.iso --cert custom.crt
#   # or
#   ./flush_crt_iso.sh --initrd initrd-xxxx --cert custom.crt
#
#   exit
#   ls ./build/
#

set -ex

BASE_DIR=/tmp
ORIGIN_DIR=/tmp/origin
NEW_DIR=/tmp/new
WORK_DIR=/tmp/work

mkdir -p ${ORIGIN_DIR} ${NEW_DIR} ${WORK_DIR} ${DAPPER_SOURCE}/build

while [ "$#" -gt 0 ]; do
    case $1 in
        --initrd)
            shift 1
            INITRD_FILE=$(readlink -f $1)
            ;;
        --iso)
            shift 1
            ISO_FILE=$(readlink -f $1)
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
mount -t iso9660 -o loop ${ISO_FILE} ${ORIGIN_DIR}
cp -rf ${ORIGIN_DIR}/* ${NEW_DIR}

# copy the initrd file
INITRD_NAME=$(basename ${ORIGIN_DIR}/boot/initrd-*)
cp ${ORIGIN_DIR}/boot/initrd-* ${WORK_DIR}/

rebuild_initrd ${INITRD_NAME} ${NEW_DIR}/boot

pushd ${NEW_DIR}
xorriso \
    -as mkisofs \
    -l -J -R -V "${DISTRIB_ID}" \
    -no-emul-boot -boot-load-size 4 -boot-info-table \
    -b boot/isolinux/isolinux.bin -c boot/isolinux/boot.cat \
    -isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin \
    -o $(basename ${ISO_FILE}) .
popd

# copy out
umount ${ORIGIN_DIR}
cp ${NEW_DIR}/$(basename ${ISO_FILE}) ${DAPPER_SOURCE}/build/
