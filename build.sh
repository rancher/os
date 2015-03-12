#!/bin/bash
set -e

cd $(dirname $0)

export DOCKER_IMAGE=rancher-os-build

./scripts/ci "$@"
mkdir -p dist
docker run --rm -it -e CHOWN_ID=$(id -u) -v $(pwd)/dist:/source/target $DOCKER_IMAGE

ARGS=`getopt -o '' -l images,push -- "$@"`
eval set -- "${ARGS}"

while true
do
    case "$1" in
        --images)
            echo "Build images"
            ./scripts/build-extra-images
            shift
            ;;
        --push)
            echo 'docker push rancher/ubuntuconsole'
            docker push rancher/ubuntuconsole
            shift
            ;;
        *)
            shift
            ;;
    esac
done

