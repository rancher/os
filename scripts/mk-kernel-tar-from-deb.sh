#!/bin/bash
set -e -x

# This script will convert an Ubuntu deb file to the kernel tar structure the RancherOS build expects
# For example 
#
#    ./scripts/mk-kernel-tar-from-deb.sh linux-image-3.19.0-28-generic_3.19.0-28.30_amd64.deb linux-image-extra-3.19.0-28-generic_3.19.0-28.30_amd64.deb linux-firmware_1.143.3_all.deb 
#

mkdir -p $(dirname $0)/../build
BUILD=$(mktemp -d $(dirname $0)/../build/deb-XXXXX)
mkdir -p $BUILD

extract()
{
    if [ ! -e $1 ]; then
        echo $1 does not exist
        exit 1
    fi

    local deb=$(readlink -f $1)

    cd $BUILD
    rm -f data.tar.* 2>/dev/null || true
    ar x $deb
    tar xvf data.tar.*
    cd -
}

for i in "$@"; do
    extract $i
done

cd $BUILD

KVER=$(ls ./lib/modules)
depmod -b . $KVER

echo Creating ${OLDPWD}/kernel.tar.gz
tar cvzf ${OLDPWD}/kernel.tar.gz ./lib boot/vmlinuz*
echo Created ${OLDPWD}/kernel.tar.gz

cd -
rm -rf ${BUILD}
