---
title: Installing RancherOS to Disk
layout: os-default
---

## Installing RancherOS to Disk
---
RancherOS comes with a simple installer that will install RancherOS on a given target disk. To install RancherOS on a new disk, you can use the `ros install` command. Before installing, you'll need to have already [booted RancherOS from ISO]({{site.baseurl}}/os/running-rancheros/workstation/boot-from-iso). Please be sure to pick the `rancheros.iso` from our release [page](https://github.com/rancher/os/releases).

### Using `ros install` to Install RancherOS

The `ros install` command orchestrates the installation from the `rancher/os` container. You will need to have already created a cloud-config file and found the target disk.

#### Cloud-Config

The easiest way to log in is to pass a `cloud-config.yml` file containing your public SSH keys. To learn more about what's supported in our cloud-config, please read our [documentation]({{site.baseurl}}/os/configuration/#cloud-config).

The `ros install` command will process your `cloud-config.yml` file specified with the `-c` flag. This file will also be placed onto the disk and installed to `/var/lib/rancher/conf/`. It will be evaluated on every boot.

Create a cloud-config file with a SSH key, this allows you to SSH into the box as the rancher user. The yml file would look like this:

```yaml
#cloud-config
ssh_authorized_keys:
  - ssh-rsa AAA...
```

<br>

You can generate a new SSH key for `cloud-config.yml` file by following this [article](https://help.github.com/articles/generating-ssh-keys/).

Copy the public SSH key into RancherOS before installing to disk.

Now that our `cloud-config.yml` contains our public SSH key, we can move on to installing RancherOS to disk!

```
$ sudo ros install -c cloud-config.yml -d /dev/sda
INFO[0000] No install type specified...defaulting to generic
Installing from rancher/os:v0.5.0
Continue [y/N]:
```

You will be prompted to see if you want to continue. Type **y**.

```
Unable to find image 'rancher/os:v0.5.0' locally
v0.5.0: Pulling from rancher/os
...
...
...
Status: Downloaded newer image for rancher/os:v0.5.0
+ DEVICE=/dev/sda
...
...
...
+ umount /mnt/new_img
Continue with reboot [y/N]:
```

After installing RancherOS to disk, you will no longer be automatically logged in as the `rancher` user. You'll need to have added in SSH keys within your [cloud-config file]({{site.baseurl}}/os/configuration/#cloud-config).

#### Installing a Different Version

By default, `ros install` uses the same installer image version as the ISO it is run from. The `-i` option specifies the particular image to install from. To keep the ISO as small as possible, the installer image is downloaded from DockerHub and used in System Docker. For example for RancherOS v0.5.0 the default installer image would be `rancher/os:v0.5.0`.

You can use `ros os list` command to find the list of available RancherOS images/versions.

```
$ sudo ros os list
rancher/os:v0.4.0 remote
rancher/os:v0.4.1 remote
rancher/os:v0.4.2 remote
rancher/os:v0.4.3 remote
rancher/os:v0.4.4 remote
rancher/os:v0.4.5 remote
rancher/os:v0.5.0 remote
```

Alternatively, you can set the installer image to any image in System Docker to install RancherOS. This is particularily useful for machines that will not have direct access to the internet.

### SSH into RancherOS

After installing RancherOS, you can ssh into RancherOS using your private key and the **rancher** user.

```
$ ssh -i /path/to/private/key rancher@<ip-address>
```

### Installing with no Internet Access

If you'd like to install RancherOS onto a machine that has no internet access, it is assumed you either have your own private registry or other means of distributing docker images to System Docker of the machine. If you need help with creating a private registry, please refer to the [Docker documentation for private registries](https://docs.docker.com/registry/).

In the installation command (i.e. `sudo ros install`), there is an option to pass in a specific image to install. As long as this image is available in System Docker, then RancherOS will use that image to install RancherOS.

```
$ sudo ros install -c cloud-config.yml -d /dev/sda -i <Image_Name_in_System_Docker>
INFO[0000] No install type specified...defaulting to generic
Installing from <Image_Name_in_System_Docker>
Continue [y/N]:
```
