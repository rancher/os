#!bin/bash

set -x -e

exec logrotate -v /etc/logrotate.conf
