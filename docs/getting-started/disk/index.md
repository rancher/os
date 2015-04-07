---
title: Getting Started on Disk
layout: default

---
# Installing to Disk

To install RancherOS on a new disk you can now use the `rancheros-install` command. 

For non-ec2 installs, before getting started create a cloud-init file that will provide your initial ssh keys. At a minimum something like:

```
#cloud-config
ssh_authorized_keys:
- ssh-rsa AAA... user@rancher
```

See section below for current supported cloud-init functionality.

The command arguments are as follows:

```
Usage:
rancheros-install [options]
Options:
-c cloud-config file
needed for SSH keys.
-d device
-f [ DANGEROUS! Data loss can happen ] partition/format without prompting
-t install-type:
generic
amazon-ebs
-v os-installer version.
-h print this
```

This command orchestrates installation from the rancher/os container. 

#### Examples:
Virtualbox installation:

`sudo rancheros-install -d /dev/sda -c ./cloud_data.yml -v v0.1.1 -t generic`





