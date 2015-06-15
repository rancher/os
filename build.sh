#!/bin/bash
set -e

cd $(dirname $0)

export DOCKER_IMAGE=rancher-os-build

source ./scripts/version

./scripts/ci

rm -rf dist

echo "Build complete. Copying artifacts..."
DIST_CONTAINER=$(docker create ${DOCKER_IMAGE})
cleanup() {
    docker rm -v ${DIST_CONTAINER}
}
trap cleanup EXIT
docker cp ${DIST_CONTAINER}:/source/dist/artifacts dist

ls -l dist/artifacts
