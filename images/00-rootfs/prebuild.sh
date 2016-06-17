#!/bin/bash

TAR=${DOWNLOADS}/rootfs.tar

if [ -e $TAR ]; then
    cd $(dirname $0)
    mkdir -p build
    cp $TAR build
fi
