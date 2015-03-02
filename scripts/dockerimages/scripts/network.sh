#!/bin/sh

/sbin/ip addr add 127.0.0.1/8 dev lo
udhcpc -i eth0
