#!/bin/bash

if [ "$DAEMON" = true ]; then
    exec udevd
fi

udevd --daemon
udevadm trigger --action=add
udevadm settle

dev=$(ros config get rancher.state.dev)
wait=$(ros config get rancher.state.wait)
if [ "$BOOTSTRAP" != true ] || [ "$dev" == "" ] || [ "$wait" != "true" ]; then
    exit
fi

for i in `seq 1 30`; do
    drive=$(ros dev $dev)
    if [ "$drive" != "" ]; then
        break
    fi
    sleep 1
done
drive=$(ros dev $dev)
if [ "$drive" = "" ]; then
    exit
fi
for i in `seq 1 30`; do
    if [ -e $drive ]; then
        break
    fi
    sleep 1
done
