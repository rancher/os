---
title: Adding kernel parameters
layout: os-default

---

## Kernel boot parameters

RancherOS parses the Linux kernel boot cmdline to add any keys it understands to its configuration. This allows you to modify what cloud-init sources it will use on boot, to enable `rancher.debug` logging, or to almost any other configuration setting.

There are two ways to set or modify persistent kernel parameters, in-place (editing the file and reboot) or during installation to disk.

### In-place editing

To edit the kernel boot parameters of an already installed RancherOS system, use the new `sudo ros config syslinux` editing command (uses `vi`).

> To activate this setting, you will need to reboot.

### During installation

If you want to set the extra kernel parameters when you are [Installing RancherOS to Disk]({{site.baseurl}}/os/running-rancheros/server/install-to-disk/) please use the `--append` parameter.

```bash
$ sudo ros install -d /dev/sda --append "rancheros.autologin=tty1"
```

### Graphical boot screen

RancherOS v1.1.0 added a Syslinux boot menu, which allows you to temporarily edit the boot paramters, or to select "Debug logging", "Autologin", both "Debug logging & Autologin" and "Recovery Console".


On desktop systems the Syslinux boot menu can be switched to graphical mode by adding `UI vesamenu.c32` to a new line in `global.cfg` (use `sudo ros config syslinux` to edit the file).

### Useful RancherOS cloud-init or boot settings

#### Recovery console

`rancher.recovery=true` will start a single user `root` bash session as easily in the boot process, with no network, or persitent filesystem mounted. This can be used to fix disk problems, or to debug your system.

#### Enable/Disable sshd

`rancher.ssh.daemon=false` (its enabled in the os-config) can be used to start your RancherOS with no sshd daemon. This can be used to further reduce the ports that your system is listening on.

#### Enable debug logging

`rancher.debug=true` will log everything to the console for debugging.

#### Autologin console

`rancher.autologin=<tty...>` will automatically log in the sepcified console - common values are `tty1`, `ttyS0` and `ttyAMA0` - depending on your platform.

#### Enable/Disable hypervisor service auto-enable

RancherOS v1.1.0 added detetion of Hypervisor, and then will try to download the a service called `<hypervisor>-vm-tools`. This may cause boot speed issues, and so can be disabled by setting `rancher.hypervisor_service=false`.
