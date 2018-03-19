#!/bin/bash
set -ex

cd $(dirname $0)/..
CD=$(pwd)

# prepare dapper needed.
rm -rf dist
mkdir -p dist/build
cp -r assets/build/initrd/ dist/build/

# prepare env needed.
DISTRO_INFO=$(cat assets/version.txt)
VERSION=$(echo -e "$DISTRO_INFO" | cut -f3 -d ' ')
COMMIT=$(echo -e "$DISTRO_INFO" | cut -f5 -d ' ')
INITRD=initrd-${VERSION}
ISO=$(echo ${DISTRIB_ID} | tr '[:upper:]' '[:lower:]')-vmware.iso
COMPRESS="xz --format=lzma -9 --memlimit-compress=80% -e"

# prepare vmware tools.
docker pull ${OS_SERVICE_VMTOOL}
docker save ${OS_SERVICE_VMTOOL} | xz -9 > dist/build/initrd/usr/share/ros/images-vmtools.tar

# package initrd.
cd dist/build/initrd/

if [ ! -f ${CD}assets/dist/artifacts/${INITRD} ]; then
    echo "Skipping package-iso vmware build: ${INITRD} not found"
    exit 1
fi

find | cpio -H newc -o | ${COMPRESS} > ${INITRD}

cd ${CD}
scp -r assets/dist/boot dist/
mkdir -p dist/rancheros
cp dist/build/initrd/${INITRD} dist/boot
cp assets/dist/artifacts/vmlinuz-${KERNEL_VERSION_amd64} dist/boot/
cp /usr/lib/ISOLINUX/isolinux.bin dist/boot/isolinux/
cp /usr/lib/syslinux/modules/bios/ldlinux.c32 dist/boot/isolinux/
cp /usr/lib/syslinux/modules/bios/*.c32 dist/boot/isolinux/
cp assets/dist/artifacts/installer.tar dist/rancheros/
cp assets/dist/artifacts/Dockerfile.amd64 dist/rancheros/

if [ -f dist/rancheros/installer.tar.gz ]; then
    rm -rf dist/rancheros/installer.tar.gz
fi

gzip -9 dist/rancheros/installer.tar
cd dist
rm -rf build/

xorriso \
    -as mkisofs \
    -l -J -R -V "${DISTRIB_ID}" \
    -no-emul-boot -boot-load-size 4 -boot-info-table \
    -b boot/isolinux/isolinux.bin -c boot/isolinux/boot.cat \
    -isohybrid-mbr /usr/lib/ISOLINUX/isohdpfx.bin \
    -o $ISO .
