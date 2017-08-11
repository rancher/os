---
title: Using ZFS in RancherOS
layout: os-default
redirect_from:
  - os/configuration/storage/
---

## Using ZFS
---

#### Installing the ZFS service

The `zfs` service will install the kernel-headers for your kernel (if you build your own kernel, you'll need to replicate this service), and then download the [ZFS on Linux]() source, and build and install it. Then it will build a `zfs-tools` image that will be used to give you access to the zfs tools.

The only restriction is that you must mount your zpool into `/mnt`, as this is the only shared mount directory that will be accessible throughout the system-docker managed containers (including the console).


```
$ sudo ros service enable zfs
$ sudo ros service up zfs
# you can follow the progress of the build by running the following command in another ssh session:
$ sudo ros service logs --follow zfs
# wait until the build is finished.
$ lsmod | grep zfs
```

> *Note:* if you switch consoles, you may need to re-run `ros up zfs`.

#### Creating ZFS pools

After it's installed, it should be ready to use. Make a zpool named `zpool1` using a device that you haven't yet partitioned (you can use `sudo fdisk -l` to list all the disks and their partitions).

> *Note:* You need to mount the zpool in `/mnt` to make it available to your host and in containers.


```
$ sudo zpool list
$ sudo zpool create zpool1 -m /mnt/zpool1 /dev/<some-disk-dev>
$ sudo zpool list
$ sudo zfs list
$ sudo cp /etc/* /mnt/zpool1
$ docker run --rm -it -v /mnt/zpool1/:/data alpine ls -la /data
```

<br>

To experiment with ZFS, you can create zpool backed by just ordinary files, not necessarily real block devices. In fact, you can mix storage devices in your ZFS pools; it's perfectly fine to create a zpool backed by real devices **and** ordinary files.

#### Using the ZFS debugger utility

The `zdb` command may be used to display information about ZFS pools useful to diagnose failures and gather statistics. By default the utility tries to load pool configurations from `/etc/zfs/zpool.cache`. Since the RancherOS ZFS service does not make use of the ZFS cache file and instead detects pools by inspecting devices, the `zdb` utility has to be invoked with the `-e` flag.

E.g. to show the configuration for the pool `zpool_1` you may run the following command:

> $ sudo zdb -e -C zpool_1


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
$ sudo ros config set rancher.docker.storage_driver 'zfs'
$ sudo ros config set rancher.docker.graph /mnt/zpool1/docker
# Now that you've changed the Docker daemon args, you'll need to start Docker
$ sudo system-docker start docker
```

After customizing the Docker daemon arguments and restarting `docker` system service, ZFS will be used as Docker storage driver:

```
$ docker info
Containers: 0
 Running: 0
 Paused: 0
 Stopped: 0
Images: 0
Server Version: 1.12.6
Storage Driver: zfs
 Zpool: error while getting pool information strconv.ParseUint: parsing "": invalid syntax
 Zpool Health: not available
 Parent Dataset: zpool1/docker
 Space Used By Parent: 19456
 Space Available: 8256371200
 Parent Quota: no
 Compression: off
Logging Driver: json-file
Cgroup Driver: cgroupfs
Plugins:
 Volume: local
 Network: host bridge null overlay
Swarm: inactive
Runtimes: runc
Default Runtime: runc
Security Options: seccomp
Kernel Version: 4.9.6-rancher
Operating System: RancherOS v0.8.0-rc8
OSType: linux
Architecture: x86_64
CPUs: 1
Total Memory: 1.953 GiB
Name: ip-172-31-24-201.us-west-1.compute.internal
ID: IEE7:YTUL:Y3F5:L6LF:5WI7:LECX:YDB5:LGWZ:QRPN:4KDI:LD66:KYTC
Docker Root Dir: /mnt/zpool1/docker
Debug Mode (client): false
Debug Mode (server): false
Registry: https://index.docker.io/v1/
Insecure Registries:
 127.0.0.0/8

```
