#!/bin/sh

/sbin/ip addr add 127.0.0.1/8 dev lo
/sbin/ip link set up dev lo
udhcpc -i eth0
