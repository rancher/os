---
title: Installing to Disk
layout: default
---

## Installing RancherOS to Disk
---
RancherOS comes with a simple installer that will install RancherOS on a given target disk. To install RancherOS on a new disk, you can use the `rancheros-install` [command]({{site.baseurl}}/docs/rancheros-tools/rancheros-install). Before installing, you'll need to have already [booted RancherOS from iso]({{site.baseurl}}/docs/running-rancheros/workstation/boot-from-iso). Please be sure to pick the `rancheros.iso` from our release [page](https://github.com/rancherio/os/releases).

### Using `rancheros-install` to Install RancherOS 

The `rancheros-install` command orchestrates the installation from the `rancher/os` container. You will need to have already created a cloud config file and found the target disk.

##### Cloud Config

The easiest way to log in is to pass a `cloud-config.yml` file containing your public SSH keys. To learn more about what's supported in our cloud-config, please read this [doc]({{site.baseurl}}/docs/cloud-config/). 

The `rancheros-install` command will process your `cloud-config.yml` file specified with the `-c` flag. This file will also be placed onto the disk and installed to `/var/lib/rancher/conf/`. It will be evaluated on every boot and be converted to `/var/lib/rancher/conf/cloud-config-processed.yml`. 

Create a cloud config file with a SSH key, this allows you to SSH into the box as the rancher user. The yml file would look like this:

```yaml
#cloud-config

ssh_authorized_keys:
  - ssh-rsa AAA... user@host
```

You can generate a new SSH key for `cloud-config.yml` file by following this [article](https://help.github.com/articles/generating-ssh-keys/). 

Copy the public SSH key into RancherOS before installing to disk. 

Now that our `cloud_config.yml` contains our public SSH key, we can move on to installing RancherOS to disk!

```bash
$ sudo rancheros-install -c cloud_config.yml -d /dev/sda 
All data will be wiped from this device
Partition: true
DEVICE: /dev/sda
Are you sure you want to continue? [yN]
```

You will be prompted to see if you want to continue. Type **y**.

```bash
Are you sure you want to continue? [yN]yUnable to find image 'rancher/os:v0.2.1
Pulling repository rancher/os
...
...
...
Downloaded newer image for rancher/os:v0.2.1
+ DEVICE=/dev/sda
...
...
...
RancherOS has been installed. Please reboot...
```
After you install RancherOS to disk, the rancher/rancher user/password will no longer be valid and you'll need to have added in SSH keys or another user within your [cloud config file]({{site.baseurl}}/docs/cloud-config/).

### SSH into RancherOS

After installing RancherOS, you can ssh into RancherOS using your private key and the **rancher** user.

```bash
$ ssh -i /path/to/private/key rancher@<ip-address>
```

