---
title: Upgrading RancherOS
layout: default

---

## Upgrading RancherOS

RancherOS has released a new version and you want to learn how to upgrade your OS. We make it easy using the [RancherCTL OS]({{site.baseurl}}/docs/rancherctl/os) commands.

RancherOS only supports upgrading from the previous version. If you are on version v0.1.2, we recommend upgrading to v0.2.0 before upgrading to v0.2.1. 

To see all of our releases, please visit our [release page](https://github.com/rancherio/os/releases) in GitHub.

### Version Control

First, let's check what version you have running on your system.

```bash
$ sudo rancherctl -v
rancherctl version v0.2.0
```

If you just want to find out the available releases from the command line, it's a simple [RancherCTL]({{site.baseurl}}/docs/rancherctl/) command.

```bash
$ sudo rancherctl os list
rancher/os:v0.1.2 remote
rancher/os:v0.2.0-rc1 remote
rancher/os:v0.2.0-rc2 remote
rancher/os:v0.2.0-rc3 remote
rancher/os:v0.2.0 local
rancher/os:v0.2.1 local
rancher/os:v0.3.0-rc1 remote
```

The local/remote shows which images are already in the system-docker containers versus on Docker Hub. If you choose to upgrade to a version that is remote, we will automatically pull that image during the upgrade.

### Upgrading 

Let's walk through upgrading! The `rancherctl os upgrade` command will automatically upgrade to the current release of RancherOS. The current release is designated as the most recent release of RancherOS. If you are a couple of versions behind, it's recommended to upgrade to each subsequent versions as RancherOS doesn't support going from v0.1.2 to v0.2.1 (skipping v0.2.0). 

```bash
$ sudo rancherctl os upgrade
Upgrading to rancher/os:v0.2.1
Continue [y/N]: 
```

Confirm that you want to continue and the final step will be to confirm that you want to reboot.

```bash
Continue [y/N: y
+ getopts c:d:t:r: OPTION
...
...
...
+ cp /dist/vmlinuz /mnt/new_img/boot/vmlinuz-v0.2.1-rancheros
Continue with reboot [y/N]: 
INFO[0005] Rebooting 
```

Confirm the reboot and your system will be rebooted. You will need to log back into your VM. After logging back in, you can check that your version has been updated.

```bash
$ sudo rancherctl -v
rancherctl version v0.2.1
```

Note: If you are booting from ISO and have not installed to disk, your upgrade will not be saved. You can view our guide to [installing to disk]({{site.baseurl}}/docs/getting-started/disk/). 

#### Upgrading to a Specific Version

If you are a couple of versions behind the current version, use the `-i` option to pick the version that you want to upgrade to. 

```bash
$ sudo rancherctl os upgrade -i rancher/os:v0.2.0
Upgrading to rancher/os:v0.2.0
Continue [y/N]: y
...
...
...
Continue with reboot [y/N]: y
INFO[0005] Rebooting 
```

#### Tips

With v0.3.0+, we have added the ability to bypass the prompts. Use the `-f` or `--force` option when upgrading. Your machine will automatically be rebooted and you'll just need to log back in when it's done.


### Rolling back an Upgrade

If you've upgraded your RancherOS and something's not working anymore, you can easily rollback your upgrade.

The `rancherctl os upgrade` command works for rolling back. We'll use the `-i` option to "upgrade" to a specific version. All you need to do is pick the previous version! Same as before, you will be prompted to confirm your upgrade version as well as confirm your reboot. 

```bash
$ sudo rancherctl -v
rancherctl version v0.2.1
$ sudo rancherctl os upgrade -i rancher/os:v0.2.0
Upgrading to rancher/os:v0.2.0
Continue [y/N]: y
...
...
...
Continue with reboot [y/N]: y
INFO[0005] Rebooting 
```
The rollback is complete and we just need to log back into our VM to confirm our rollback.

```bash
$ sudo rancherctl -v
rancherctl version 0.2.0
```

### Staging an Upgrade

During an upgrade, the template of the upgrade is downloaded from the rancher/os repository. You can download this template ahead of time so that it's saved locally. This will decrease the time it takes to upgrade. We'll use the `-s` option to stage the specific template. You will need to specify the image name with the `-i` option, otherwise it will automatically stage the current version. 

```bash
$ sudo rancherctl os upgrade -s -i rancher/os:v0.2.1
Pulling image (v0.2.1) from rancher/os
Pulling image (v0.2.1) from rancher/os, endpoint: https://registry-1.docker.io/v1/
...
...
...
Status: Downloaded newer image for rancher/os:v0.2.1
```







