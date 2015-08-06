#!/bin/bash
set -e

docker rm -fv ros-build > /dev/null 2>&1 || :
exec docker run -v /var/run/docker.sock:/var/run/docker.sock --name=ros-build ros-build "$@"
