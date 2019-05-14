#!/bin/sh
set -e -x

root="/mnt/install"
cache="/var/lib/rancher/cache"
preload="/var/lib/rancher/preload/system-docker"

partition=$1
images=$2

cache_services() {
    mkdir -p ${root}${cache}
    cp ${cache}/* ${root}${cache}
}

cache_images() {
    mkdir -p ${root}${preload}
    for i in ${images}; do
        system-docker pull $i
    done
    system-docker save -o ${root}${preload}/os-include.tar ${images}
}

mkdir -p ${root}
mount ${partition} ${root}
cache_services
cache_images
umount ${root}
