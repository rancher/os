#!/bin/bash

if [ ! -d "/sys/class/net/eth0" ]; then
    echo "eth0 not found, ssh access cannot be provided!"
fi

udhcpc -i eth0
