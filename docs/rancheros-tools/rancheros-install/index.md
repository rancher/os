---
title: rancheros-install
layout: default
---

## rancheros-install
---

The `rancheros-install` command is used to install RancherOS to hard disk. Please follow the [guide]({{site.baseurl}}/docs/running-rancheros/server/install-to-disk/) for an example of how to install to disk.

### Options

| Options | Description |
|--------|---------|
|-c | The Cloud-Config file needed for SSH keys |
| -d | Which Device to Install to |
|    -f | [ DANGEROUS! Data loss can happen ] Partition/Format without prompting |
|    -t | Decide the install-type: generic:    (Default) Creates 1 ext4 partition and installs RancherOS amazon-ebs: Installs RancherOS and sets up PV-GRUB
|    -v | Provide which os-installer version |
|    -h | Prints the help informations |

#### Cloud-Config

The `-c` option provides the location of the cloud config file. Read more about [cloud config files]({{site.baseurl}}/docs/cloud-config)

#### Device

The `-d` option provide the target disk location. 

You can see the list of disks available to install to by running `sudo fdisk -l`

```bash
[rancher@rancher ~]$ sudo fdisk -l
Disk /dev/sda: 8589 MB, 8589934592 bytes
255 heads, 63 sectors/track, 1044 cylinders
Units = cylinders of 16065 *512 = 8225280 bytes

Disk /dev/sda doesn't contain a valid partition table
[rancher@rancher ~]$
```

#### Install Type
The `-t` option determines what type of installation is used. The _amazon-ebs_ type is for creating your own AMI images. Since we are creating the [RancherOS AMI images]({{site.baseurl}}/docs/running-rancheros/cloud/aws/), there is no need to create your own. 

By default, the install type will be generic.

#### Version

The `-v` option will indicate which version of RancherOS to install. To see the most recent versions of RancherOS, please vist the RancherOS GitHub page of [releases](https://github.com/rancherio/os/releases).

By default, the version installed will be version that RancherOS is currently running.

You can use the [rancherctl os]({{site.baseurl}}/docs/rancheros-tools/rancherctl/os/) commands to find the list of available versions.

```bash
[rancher@rancher ~]$ sudo rancherctl os list
rancher/os:v0.1.2 remote
rancher/os:v0.2.0-rc1 remote
rancher/os:v0.2.0-rc2 remote
rancher/os:v0.2.0-rc3 remote
rancher/os:v0.2.0 remote
rancher/os:v0.2.1 remote
rancher/os:v0.3.0-rc1 remote
[rancher@rancher ~]$ 
```

