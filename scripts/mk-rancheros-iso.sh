#!/bin/bash
set -ex

cd $(dirname $0)/..
. scripts/build-common

mkdir -p ${CD}/boot/isolinux

cp ${DIST}/artifacts/initrd                   ${CD}/boot
cp ${DIST}/artifacts/vmlinuz                  ${CD}/boot
cp scripts/isolinux.cfg                       ${CD}/boot/isolinux
cp /usr/lib/ISOLINUX/isolinux.bin             ${CD}/boot/isolinux
cp /usr/lib/syslinux/modules/bios/ldlinux.c32 ${CD}/boot/isolinux
cd ${CD} && xorriso \
    -publisher "Rancher Labs, Inc." \
    -as mkisofs \
    -l -J -R -V "RancherOS" \
    -no-emul-boot -boot-load-size 4 -boot-info-table \
    -b boot/isolinux/isolinux.bin -c boot/isolinux/boot.cat \
    -isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin \
    -o ${DIST}/artifacts/rancheros.iso ${CD}
