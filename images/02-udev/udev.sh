#!/bin/bash

if [ "$DAEMON" = true ]; then
    exec udevd
fi

udevd --daemon
udevadm trigger --action=add
udevadm settle

if [ "$BOOTSTRAP" = true ]; then
    # This was needed to get USB devices to fully register
    # There is probably a better way to do this
    killall udevd
    udevd --daemon
    udevadm trigger --action=add
    udevadm settle
fi
