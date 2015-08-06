#!/bin/bash
set -e

DOCKER_ARGS=
if [ -n "$BIND_DIR" ]; then
    if [ "$BIND_DIR" = "." ]; then
        BIND_DIR=$(pwd)
    fi
    DOCKER_ARGS="-t -v $BIND_DIR:/go/src/github.com/rancherio/os"
fi

docker rm -fv ros-build >/dev/null 2>&1 || true
exec docker run -i -v /var/run/docker.sock:/var/run/docker.sock $DOCKER_ARGS --name=ros-build ros-build "$@"
