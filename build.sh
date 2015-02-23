#!/bin/bash
set -e

cd $(dirname $0)

export DOCKER_IMAGE=rancher-os-build

./scripts/ci "$@"
mkdir -p dist
docker run --rm -it -e CHOWN_ID=$(id -u) -v $(pwd)/dist:/source/target $DOCKER_IMAGE

# Stupidest argparse ever
if echo "$@" | grep -q -- '--images'; then
    ./scripts/build-extra-images
    echo 'docker push rancher/ubuntuconsole'
fi

# And again
if echo "$@" | grep -q -- '--push'; then
    docker push rancher/ubuntuconsole
fi
