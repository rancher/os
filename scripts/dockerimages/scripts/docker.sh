#!/bin/sh
set -x -e

CGROUPS="perf_event net_cls freezer devices blkio memory cpuacct cpu cpuset"

mkdir -p /sys/fs/cgroup
mount -t tmpfs none /sys/fs/cgroup

for i in $CGROUPS; do
    mkdir -p /sys/fs/cgroup/$i
    mount -t cgroup -o $i none /sys/fs/cgroup/$i
done

if ! lsmod | grep -q br_netfilter; then
    modprobe br_netfilter
fi

rm -f /var/run/docker.pid
exec docker -d -s overlay
