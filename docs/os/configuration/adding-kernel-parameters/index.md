---
title: Adding kernel parameters
layout: os-default

---

## Kernel parameters

There are two ways to edit the kernel parameters, in-place (editing the file and reboot) or during installation to disk.

### In-place editing

For in-place editing, you will need to run a container with an editor and a mount to access the `/boot/global.cfg` file containing the kernel parameters.

> To activate this setting, you will need to reboot.

```bash
$ sudo system-docker run --rm -it -v /:/host alpine vi /host/boot/global.cfg
```


### During installation

If you want to set the extra kernel parameters when you are [Installing RancherOS to Disk]({{site.baseurl}}/os/running-rancheros/server/install-to-disk/) please use the `--append` parameter.

```bash
$ sudo ros install -d /dev/sda --append "rancheros.autologin=tty1"
```
