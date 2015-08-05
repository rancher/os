#!/bin/bash
set -ex

cd $(dirname $0)/..
. scripts/build-common

cd ${DIST}/artifacts
rm -f iso-checksums.txt || :

for algo in 'sha256' 'md5'; do
    echo "$algo: `${algo}sum rancheros.iso`" >> iso-checksums.txt;
done
