#!/bin/bash
set -e

DOCKER_ARGS=
if [ -n "$BIND_DIR" ]; then
    if [ "$BIND_DIR" = "." ]; then
        BIND_DIR=$(pwd)
    fi
    DOCKER_ARGS="-t -v $BIND_DIR:/go/src/github.com/rancherio/os"
fi
if [ -c /dev/kvm ] || [ "${KVM}" == "1" ]; then
    DOCKER_ARGS="${DOCKER_ARGS} --device=/dev/kvm:/dev/kvm"
fi

NAME=ros-build
while [ "$#" -gt 0 ]; do
    case $1 in
        --name)
            shift 1
            NAME="$1"
            ;;
        --rm)
            NAME=$(mktemp ${NAME}-XXXXXX)
            rm $NAME
            DOCKER_ARGS="${DOCKER_ARGS} --rm"
            ;;
        *)
            break
            ;;
    esac
    shift 1
done

DOCKER_ARGS="${DOCKER_ARGS} --name=${NAME}"
docker rm -fv ${NAME} >/dev/null 2>&1 || true
exec docker run -i -v /var/run/docker.sock:/var/run/docker.sock $DOCKER_ARGS ros-build "$@"
