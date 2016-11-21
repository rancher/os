---
title: Using ZFS in RancherOS
layout: os-default
redirect_from:
  - os/configuration/storage/
---

## Using ZFS

In order to start using ZFS, you'll need to first enable one of the [persistent consoles]({{site.baseurl}}/os/configuration/custom-console/#console-persistence) and enable [kernel headers]({{site.baseurl}}/os/configuration/kernel-modules-kernel-headers/).

```
$ sudo ros console switch ubuntu
$ sudo ros service enable kernel-headers
$ sudo ros service up kernel-headers
$ sudo ros service logs kernel-headers
```

When RancherOS console has reloaded, you will have logged into the persistent console. The current kernel headers will need to be downloaded using the `ros service enable` and the service will be started with `ros service up kernel-headers`.

You can make sure that the headers are in the right place by looking at the kernel-headers service logs.

#### Installing ZFS on Ubuntu Console

Based on the [Ubuntu ZFS docs](https://wiki.ubuntu.com/Kernel/Reference/ZFS), you only need to install `zfs` package into the Ubuntu console to enable ZFS. All the other necessary packages will be installed as its dependencies.

```
$ sudo apt update
$ sudo apt install zfs
```

Then have a look at the zfs dmks build log file - which is likely to be a file similar to `/var/lib/dkms/zfs/0.6.5.6/build/make.log`.

#### Mounting ZFS filesystems on boot

In order for ZFS to load on boot, it needs to be added to `modules` list in the config. Prior to adding it to the list of modules, you'll need to check to see if there are other modules that are currently enabled.

```
# Check to see what modules currently exist
$ sudo ros config get rancher.modules
# Make sure to include any modules that were already enabled
$ sudo ros config set rancher.modules [zfs]
```

<br>

You will also need to have the zpool cache imported on boot:

```
[ -f /etc/zfs/zpool.cache ] && zpool import -c /etc/zfs/zpool.cache -a
```

<br>

A cloud-config `runcmd` instruction will do it for you:

```
# check current 'runcmd' list
$ sudo ros config get runcmd
[]
# add the command we need to run on boot
$ sudo ros config set runcmd "[[sh, -c, '[ -f /etc/zfs/zpool.cache ] && zpool import -c /etc/zfs/zpool.cache -a']]"
```

#### Using ZFS

After it's installed, it should be ready to use!

```
$ sudo modprobe zfs
$ sudo zpool list
$ sudo zpool create zpool1 /dev/<some-disk-dev>
```

<br>

To experiment with ZFS, you can create zpool backed by just ordinary files, not necessarily real block devices. In fact, you can mix storage devices in your ZFS pools; it's perfectly fine to create a zpool backed by real devices **and** ordinary files.

## ZFS storage for Docker on RancherOS

First, you need to stop  the`docker` system service and wipe out `/var/lib/docker` folder:

```
$ sudo system-docker stop docker
$ sudo rm -rf /var/lib/docker/*
```

To enable ZFS as the storage driver for Docker, you'll need to create a ZFS filesystem for Docker and make sure it's mounted.

```
$ sudo zfs create zpool1/docker
$ sudo zfs list -o name,mountpoint,mounted
```

At this point you'll have a ZFS filesystem created and mounted at `/zpool1/docker`. According to [Docker ZFS storage docs](https://docs.docker.com/engine/userguide/storagedriver/zfs-driver/), if the Docker root dir is a ZFS filesystem, the Docker daemon will automatically use `zfs` as its storage driver.

Now you'll need to remove `-s overlay` (or any other storage driver) from the Docker daemon args to allow docker to automatically detect `zfs`.

```
$ sudo ros config set rancher.docker.storage_driver ''
$ sudo ros config set rancher.docker.graph /zpool1/docker
# After editing Docker daemon args, you'll need to start Docker
$ sudo system-docker stop docker
$ sudo system-docker start docker
```

After customizing the Docker daemon arguments and restarting `docker` system service, ZFS will be used as Docker storage driver:

```
$ docker info
Containers: 1
 Running: 0
 Paused: 0
 Stopped: 1
Images: 1
Server Version: 1.12.1
Storage Driver: zfs
 Zpool: zpool1
 Zpool Health: ONLINE
 Parent Dataset: zpool1/docker
 Space Used By Parent: 27761152
 Space Available: 4100088320
 Parent Quota: no
 Compression: off
Logging Driver: json-file
Cgroup Driver: cgroupfs
Plugins:
 Volume: local
 Network: host null bridge overlay
Swarm: inactive
Runtimes: runc
Default Runtime: runc
Security Options: seccomp
Kernel Version: 4.4.16-rancher
Operating System: RancherOS v0.6.0-rc8
OSType: linux
Architecture: x86_64
CPUs: 2
Total Memory: 1.938 GiB
Name: rancher
ID: EK7Q:WTBH:33KR:UCRY:YAPI:N7RX:D25K:S7ZH:DRNY:ZJ3J:25XE:P3RF
Docker Root Dir: /zpool1/docker
Debug Mode (client): false
Debug Mode (server): false
Registry: https://index.docker.io/v1/
Insecure Registries:
 127.0.0.0/8
```
