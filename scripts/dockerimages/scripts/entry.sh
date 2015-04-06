#!/bin/bash

if [ -e /host/dev ]; then
    mount --bind /host/dev /dev
fi

exec "$@"
