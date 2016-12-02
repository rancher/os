---
title: Running RancherOS on Raspberry Pi
layout: os-default
---

## Raspberry Pi
---

As of v0.5.0, RancherOS releases include a Raspberry Pi image that can be found on our [releases page](https://github.com/rancher/os/releases). The official Raspberry Pi documentation contains instructions on how to [install operating system images](https://www.raspberrypi.org/documentation/installation/installing-images/).

When installing, there is no ability to pass in a [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config). You will need to boot up, change the configuration and then reboot to apply those changes. 

Currently, only Raspberry Pi 2 and 3 are tested and known to work.

### Using the entire SD Card

RancherOS does not currently expand the root partition to fill the remainder of the SD card automatically. Instead, the following workaround can be used to store Docker containers on a larger partition that fills the remainder.

1. `sudo fdisk /dev/mmcblk0`
2. Create a `n`ew partition
3. Press `[Enter]` four (4x) times to accept the defaults
4. Then `w`rite the table and exit
5. `sudo reboot` to reboot and reload the new partition table
6. `sudo mkdir /mnt/docker` to create the directory to be used as the new Docker root
7. `sudo ros config set rancher.docker.extra_args [-g,/mnt/docker]` to configure Docker to use the new root
8. `sudo mkfs.ext4 /dev/mmcblk0p3` to format the disk
9. `sudo ros config set mounts "[['/dev/mmcblk0p3','/mnt/docker','ext4','']]"` to preserve this mount after reboots
10. `sudo mount /dev/mmcblk0p3 /mnt/docker` to mount the Docker root
11. `sudo system-docker restart docker` to restart Docker using the new root

If this is not a new installation, you'll have to copy over your existing Docker root (`/var/lib/docker`) to the new root (`/mnt/docker`).

1. `sudo cp -R /var/lib/docker/* /mnt/docker` to recursively copy all files
2. `sudo system-docker restart docker` to restart Docker using the new root
