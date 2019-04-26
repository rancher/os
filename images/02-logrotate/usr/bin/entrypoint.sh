#!/bin/bash

for f in /usr/share/logrotate/logrotate.d/*; do
    target=/etc/logrotate.d/$(basename ${f})
    if [ ! -e ${target} ]; then
        cp ${f} ${target}
    fi
done

exec /usr/bin/ros entrypoint "$@"
