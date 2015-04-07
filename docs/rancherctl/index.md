---
title: RancherCTL
layout: default
---

### Using RancherCTL
Another useful command that can be used with RancherOS is rancherctl which can be used to control and configure the system:

```sh
[rancher@rancher ~]$ rancherctl -v
rancherctl version 0.0.1
```

RancherOS state is controlled by simple document, rancherctl is used to edit the configuration of the system, to see for example the dns configuration of the system:

```sh
[rancher@rancher ~]$ sudo rancherctl config get dns
- 8.8.8.8
- 8.8.4.4
```

You can use rancherctl to customize the console and replace the native Busybox console with the consoles from other Linux distributions.  Initially RancherOS only supports the Ubuntu console, but other console support will be coming soon To enable the Ubuntu console use the following command:

```sh
[rancher@rancher ~]$ sudo rancherctl addon enable ubuntu-console;
[rancher@rancher ~]$ sudo reboot
```

After that you will be able to use Ubuntu console, to turn it off use disable instead of enable, and then reboot.

```sh
rancher@rancher:~$ sudo rancherctl addon disable ubuntu-console;
```

Note that any changes to the console or the system containers will be lost after reboots, any changes to /home or /opt will be persistent. Theconsole always executes **/opt/rancher/bin/start.sh** at each startup. 