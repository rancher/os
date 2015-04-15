---
title: RancherCTL OS
layout: default

---

## RancherCTL OS

`rancherctl os` gives more information on the OS versions available for RancherOS. 

Remember all `rancherctl` commands needs to be used with `sudo`. 


### Sub Commands

|Command | Description |
|--------|-------------|
|`upgrade` |	Upgrade to the latest version |
|`list`	|	List all the current available versions |
| `version` | Show the currently installed version|


### Version (Added as of v0.3.0)

Besides using `rancherctl -v`, you can also find the current version with the `version` command. 

```bash
$ sudo rancherctl os version
rancher/os:v0.2.1
```

### List

It's easy to see all the current available versions using the `list` command. 

```bash
$ sudo rancherctl os list
rancher/os:v0.1.2 remote
rancher/os:v0.2.0-rc1 remote
rancher/os:v0.2.0-rc2 remote
rancher/os:v0.2.0-rc3 remote
rancher/os:v0.2.0 local
rancher/os:v0.2.1 remote
rancher/os:v0.3.0-rc1 remote
```

### Upgrade

Our [upgrading]({{site.baseurl}}/docs/upgrading/) page provides more details on steps for upgrading.

#### Upgrade Options

|Options | Description |
|--------|-------------|
|`--stage`, `-s`	|Only stage the new upgrade, don't apply it|
|`--image`, `-i` 	|Upgrade to a certain image|
|  `--force`, `-f` |	do not prompt for input|


#### Image

The `-i` or `--image` option uses an image name to upgrade to the specific version. If no image is specified, the command will automatically use the current version.

```bash
$ sudo rancherctl os upgrade -i rancher/os:v0.2.0    
Pulling repository rancher/os
Pulling image (v0.2.0) from rancher/os
Pulling image (v0.2.0) from rancher/os, endpoint: https://registry-1.docker.io/v1/
...
...
...
Download complete
Status: Downloaded newer image for rancher/os:v0.2.0
Upgrading to rancher/os:v0.2.0
Continue [y/N]: y
Continue with reboot [y/N] y
INFO[0004] Rebooting 
```

#### Stage

The `-s` or `--stage` option is used to stage the specific template instead of immediately installing it. 

```bash
$ sudo rancherctl os upgrade -s -i rancher/os:v0.2.1
Pulling image (v0.2.1) from rancher/os
Pulling image (v0.2.1) from rancher/os, endpoint: https://registry-1.docker.io/v1/
...
...
...
Status: Downloaded newer image for rancher/os:v0.2.1
[rancher@rancher ~]$
```


#### Force (available in version v0.3.0+)

The `-f` or `--force` option allows the user to bypass the prompts for continuing with the upgrade and rebooting. It will automatically do the upgrade and reboot.

<br>



