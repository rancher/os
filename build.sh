#!/bin/bash
set -e

cd $(dirname $0)

./scripts/addon-deps

export DOCKER_IMAGE=rancher-os-build

source ./scripts/version

./scripts/ci
mkdir -p dist
docker run --rm -it -e CHOWN_ID=$(id -u) -v $(pwd)/dist:/source/target $DOCKER_IMAGE

# Stupidest argparse ever
if echo "$@" | grep -q -- '--images'; then
    ./scripts/build-extra-images
fi

# And again
if echo "$@" | grep -q -- '--push'; then
    docker push rancher/ubuntuconsole:${VERSION}
fi

ls -l dist/artifacts
