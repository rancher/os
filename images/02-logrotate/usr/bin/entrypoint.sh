#!/bin/bash

cp /usr/share/logrotate/logrotate.d/* /etc/logrotate.d

exec /usr/bin/ros entrypoint "$@"
