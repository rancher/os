#!/bin/bash
set -e

DIR=$(readlink /lib/modules/$(uname -r)/build)
STAMP=${DIR}/.done
VER=$(basename $DIR)
URL=https://github.com/rancher/os-kernel/releases/download/${VER}/build.tar.gz

if [ -e $STAMP ]; then
    echo Headers already installed in $DIR
    exit 0
fi

echo Downloading $URL
mkdir -p $DIR
wget -O - $URL | gzip -dc | tar xf - -C $DIR
touch $STAMP

echo Headers installed at $DIR
