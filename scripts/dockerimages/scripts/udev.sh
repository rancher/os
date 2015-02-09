#!/bin/sh

mount --bind /host/dev /dev
udevd --daemon
udevadm trigger --action=add
udevadm settle
