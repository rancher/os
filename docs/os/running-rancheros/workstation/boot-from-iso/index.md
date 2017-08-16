---
title: Booting from ISO

---

## Boot from ISO
---
The RancherOS ISO file can be used to create a fresh RancherOS install on KVM, VMware, VirtualBox, or bare metal servers. You can download the `rancheros.iso` file from our [releases page](https://github.com/rancher/os/releases/).

You must boot with at least **512MB** of memory. If you boot with the ISO, you will automatically be logged in as the `rancher` user. Only the ISO is set to use autologin by default. If you run from a cloud or install to disk, SSH keys or a password of your choice is expected to be used.

> **Note:** If you are planning on [installing to disk]({{page.osbaseurl}}/running-rancheros/server/install-to-disk/), you will need at least 1.5GB of RAM.

### Install to Disk

After you boot RancherOS from ISO, you can follow the instructions [here]({{page.osbaseurl}}/running-rancheros/server/install-to-disk/) to install RancherOS to a hard disk.

### Persisting State

If you are running from the ISO, RancherOS will be running from memory. All downloaded Docker images, for example, will be stored in a ramdisk and will be lost after the server is rebooted. You can
create a file system with the label `RANCHER_STATE` to instruct RancherOS to use that partition to store state. Suppose you have a disk partition on the server called `/dev/sda`, the following command formats that partition and labels it `RANCHER_STATE`

```
$ sudo mkfs.ext4 -L RANCHER_STATE /dev/sda
# Reboot afterwards in order for the changes to start being saved.
$ sudo reboot
```

After you reboot, the server RancherOS will use `/dev/sda` as the state partition.

> **Note:** If you are installing RancherOS to disk, you do not need to run this command.
