---
title: FAQs
layout: default

---

## Frequently Asked Questions

**What is required?**

Docker 1.5+ is required. 


**What are some commands?**

Command | Description
--------|------------
`docker`| Good old Docker, use that to run stuff.
`system-docker` | The docker instance running the system containers.  Must run as root or using `sudo`
`rancherctl` | Control and configure RancherOS

<br>
**How can I customize the console?**

Since RancherOS is so small the default console is based off of Busybox.  This it not always the best experience.  The intention with RancherOS is to allow you to swap out different consoles with something like Ubuntu, Fedora, or CentOS.  Currently we have Ubuntu configured but we will add more.  To enable the Ubuntu console do the following.

sudo rancherctl addon enable ubuntu-console

Run the above but with `disable` to turn it off.  Currently you have to reboot the system to enable the new console.  In the future it will be dynamic and just require you to log out and back in.

**Why are my changes to the console being lost?**

The console (and all system containers) are ephemeral.  This means on each reboot of the system all changes to the console are lost.  Any changes in `/home` or `/opt` will be persisted though.  Additionally, on startup of the console container, if `/opt/rancher/bin/start.sh` exists, it will be executed.  You can add anything to that script to configure your console the way you want it.

In the future we will allow one to provide a custom image for the console container, but we just haven't gotten around yet to enabling that.


**How do I start developing?**

Development is easiest done with QEMU on Linux.  If you aren't running Linux natively then we recommend you run VMware Fusion/Workstation and enable VT-x support.  Then, QEMU (with KVM support) will run sufficiently fast inside a Linux VM.

First run `./build.sh` to create the initial bootstrap Docker images.  After that if you make changes to the go code only run `./scripts/build`.  To launch RancherOS in QEMU from your dev version run `./scripts/run`.  You can SSH in using `ssh -l rancher -p 2222 localhost`.  Your SSH keys should have been populated so you won't need a password.  If you don't have SSH keys then the password is "rancher".