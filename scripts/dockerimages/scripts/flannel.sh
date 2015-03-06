#!/bin/sh

exec /flannel --etcd-endpoints="172.17.7.101:4001" --iface=eth1
