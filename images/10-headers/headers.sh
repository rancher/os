#!/bin/bash
set -e

DIR=$(readlink /lib/modules/$(uname -r)/build)
STAMP=${DIR}/.done
VER=$(basename $DIR)

if [ -e $STAMP ]; then
    echo Headers already installed in $DIR
    exit 0
fi

mkdir -p $DIR
cat /build.tar.gz | gzip -dc | tar xf - -C $DIR
touch $STAMP

echo Headers installed at $DIR
