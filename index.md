---
title: RancherOS Documentation
layout: default

---

## RancherOS

The smallest, easiest way to run Docker in production at scale.  Everything in RancherOS is a container managed by Docker.  This includes system services such as udev and rsyslog.  RancherOS includes only the bare minimum amount of software needed to run Docker.  This keeps the binary download of RancherOS to about 20MB.  Everything else can be pulled in dynamically through Docker.

### How this works



Everything in RancherOS is a Docker container.  We accomplish this by launching two instances of Docker.  One is what we call the system Docker which runs as PID 1.  System Docker then launches a container that runs the user Docker.  The user Docker is then the instance that gets primarily used to create containers.  We created this separation because it seemed logical and also it would really be bad if somebody did 
`docker rm -f $(docker ps -qa)` and deleted the entire OS.



![How it works]({{site.baseurl}}/img/rancheroshowitworks.png "How it works")



### Running RancherOS

To find out more about installing RancherOS, read more about it on our [Getting Started Guide]({{site.baseurl}}/docs/getting-started/).


### Latest Release

v0.2.1 - Docker 1.5.0 - Linux 3.19.2


### Building

Docker 1.5+ required.

./build.sh

When the build is done, the ISO should be in `dist/artifacts`.

### Developing

Development is easiest done with QEMU on Linux.  If you aren't running Linux natively then we recommend you run VMware Fusion/Workstation and enable VT-x support.  Then, QEMU (with KVM support) will run sufficiently fast inside a Linux VM.

First run `./build.sh` to create the initial bootstrap Docker images.  After that if you make changes to the go code only run `./scripts/build`.  To launch RancherOS in QEMU from your dev version, run `./scripts/run`.  You can SSH in using `ssh -l rancher -p 2222 localhost`.  Your SSH keys should have been populated so you won't need a password.  If you don't have SSH keys then the password is "rancher".


