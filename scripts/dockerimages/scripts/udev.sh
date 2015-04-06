#!/bin/bash

if [ "$DAEMON" = true ]; then
    exec udevd
fi

udevd --daemon
udevadm trigger --action=add
udevadm settle
