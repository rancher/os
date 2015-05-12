---
title: FAQs
layout: default

---

## Frequently Asked Questions
---

###What is required?

Docker 1.5+ is required. 


###What are some commands?

Command | Description
--------|------------
`docker`| Good old Docker, use that to run stuff.
`system-docker` | The docker instance running the system containers.  Must run as root or using `sudo`
`ros` | Control and configure RancherOS


### How can I customize the console?

Since RancherOS is so small the default console is based off of Busybox.  This it not always the best experience.  The intention with RancherOS is to allow you to swap out different consoles with something like Ubuntu, Fedora, or CentOS.  Currently we have Ubuntu configured but we will add more.  

**With v0.3.0+**, we have updated to use the `service` command within `ros` and removed `addon`. 
**With v0.3.1+**, we have updated to use `ros` instead of `rancherctl`.

To enable the Ubuntu console do the following.

```bash
[rancher@rancher ~]$ sudo ros service enable ubuntu-console;
[rancher@rancher ~]$ sudo reboot
```

Run the above but with `disable` to turn it off.  Currently, you have to reboot the system to enable the new console.  In the future, it will be dynamic and just require you to log out and back in.

```bash
[rancher@rancher ~]$ sudo ros service disable ubuntu-console;
[rancher@rancher ~]$ sudo reboot
```

**Versions prior to v0.3.0**, can follow these directions for the `addon` command.

To enable the Ubuntu console do the following.

```bash
[rancher@rancher ~]$ sudo ros addon enable ubuntu-console;
[rancher@rancher ~]$ sudo reboot
```

Run the above but with `disable` to turn it off.  Currently, you have to reboot the system to enable the new console.  In the future, it will be dynamic and just require you to log out and back in.

```bash
[rancher@rancher ~]$ sudo ros addon disable ubuntu-console;
[rancher@rancher ~]$ sudo reboot
```

### Why are my changes to the console being lost?

The console and all system containers are ephemeral.  This means on each reboot of the system all changes to the console are lost.  Any changes in `/home` or `/opt` will be persisted though.  Additionally, on startup of the console container, if `/opt/rancher/bin/start.sh` exists, it will be executed.  You can add anything to that script to configure your console the way you want it.

In the future, we will allow one to provide a custom image for the console container, but we just haven't gotten around yet to enabling that.
