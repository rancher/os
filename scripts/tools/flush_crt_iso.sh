#!/bin/bash

# How to use:
#   make shell-bind
#   cd scripts/tools/
#   wget https://link/rancheros-xxx.iso
#   wget http://link/custom.crt
#   ./flush_crt_iso.sh --iso rancheros-vmware-autoformat.iso --cert custom.crt
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

# copy the iso content
mount -t iso9660 -o loop ${ISO_FILE} ${ORIGIN_DIR}
cp -rf ${ORIGIN_DIR}/* ${NEW_DIR}

# copy the initrd file
INITRD_NAME=$(basename ${ORIGIN_DIR}/boot/initrd-*)
cp ${ORIGIN_DIR}/boot/initrd-* ${WORK_DIR}/

# update and rebuild the initrd
pushd ${WORK_DIR}
mv initrd-* ${INITRD_NAME}.gz
gzip -d ${INITRD_NAME}.gz
cpio -i -F ${INITRD_NAME}
rm -f ${INITRD_NAME}

cat ${CERT_FILE} >> ${WORK_DIR}/usr/etc/ssl/certs/ca-certificates.crt

find | cpio -H newc -o | gzip -9 > ${NEW_DIR}/boot/${INITRD_NAME}
popd

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
