#!/bin/bash

if [ -n "$1" ]; then
    exec mkfs.ext4 -L RANCHER_STATE $1
fi
