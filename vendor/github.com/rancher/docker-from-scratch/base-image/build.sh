#!/bin/bash
set -e

cd $(dirname $0)

IMAGE=${IMAGE:-dockerscratch-files}

docker build -t $IMAGE .

if [ -z "$NO_BIND" ] && [ "$(uname)" == "Linux" ]; then
    mkdir -p cache
    ARGS="-v $(pwd):/source -u $(id -u) -e HOME=/root -v $(pwd)/cache:/root"
fi

ID=$(docker run -itd $ARGS $IMAGE /source/scripts/build)
trap "docker rm -fv $ID" exit

docker attach $ID
docker wait $ID

mkdir -p dist
docker cp $ID:/source/dist/base-files.tar.gz dist

echo Done
