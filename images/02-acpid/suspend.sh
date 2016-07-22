#!/bin/sh
if [ "$3" = "close" ]; then
    echo -n "mem" > /sys/power/state
fi
