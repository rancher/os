#!/bin/bash
set -e

cd $(dirname $0)

rm -rf ./build
mkdir -p ./build
cp ./../01-base/entry.sh ./build/
cp ./../02-console/console.sh ./build/
cp ./../02-console/docker-init ./build/
cp ./../02-console/update-ssh-keys ./build/
cp ./../02-console/rancheros-install ./build/
