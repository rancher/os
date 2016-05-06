#!/bin/bash
set -e

DIR=$(readlink /lib/modules/$(uname -r)/build)
STAMP=/lib/modules/$(uname -r)/.extra-done
VER=$(basename $DIR)
URL=https://github.com/rancher/os-kernel/releases/download/${VER}/extra.tar.gz

if [ -e $STAMP ]; then
    echo Kernel extras already installed in $DIR
    exit 0
fi

echo Downloading $URL
wget -O - $URL | gzip -dc | tar xf - -C /
touch $STAMP

echo Kernel extras installed at $DIR
