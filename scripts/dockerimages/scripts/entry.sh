#!/bin/bash

if [ -e /host/dev ]; then
    mount --rbind /host/dev /dev
fi

exec "$@"
