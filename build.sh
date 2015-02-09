#!/bin/bash
set -e

cd $(dirname $0)

export DOCKER_IMAGE=rancher-os-build

./scripts/ci
mkdir -p dist
docker run -it -e CHOWN_ID=$(id -u) -v $(pwd)/dist:/source/target $DOCKER_IMAGE
