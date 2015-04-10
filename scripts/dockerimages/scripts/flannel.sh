#!/bin/sh

exec /flannel --etcd-endpoints="http://172.17.7.101:4001" --iface=eth1
