---
title: Adding kernel parameters
layout: os-default

---

## Kernel parameters

There are two ways to edit the kernel parameters, in-place (editing the file and reboot) or during installation to disk.

### In-place editing

To edit the kernel boot parameters of an already installed RancherOS system, use the new `sudo ros config syslinux` editing command (uses `vi`).

> To activate this setting, you will need to reboot.

#### Graphical boot screen

RancherOS v1.1.0 added a syslinux boot menu, which on desktop systems can be switched to graphical mode by adding `UI vesamenu.c32` to a new line in `global.cfg` (use `sudo ros config syslinux` to edit the file).

### During installation

If you want to set the extra kernel parameters when you are [Installing RancherOS to Disk]({{site.baseurl}}/os/running-rancheros/server/install-to-disk/) please use the `--append` parameter.

```bash
$ sudo ros install -d /dev/sda --append "rancheros.autologin=tty1"
```
