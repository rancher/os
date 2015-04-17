---
title: RancherOS Documentation
layout: default

---

## Contributing to RancherOS
---

### Developing

Development is easiest done with QEMU on Linux.  If you aren't running Linux natively then we recommend you run VMware Fusion/Workstation and enable VT-x support.  Then, QEMU (with KVM support) will run sufficiently fast inside a Linux VM.

First run `./build.sh` to create the initial bootstrap Docker images.  After that if you make changes to the go code only run `./scripts/build`.  To launch RancherOS in QEMU from your dev version, run `./scripts/run`.  You can SSH in using `ssh -l rancher -p 2222 localhost`.  Your SSH keys should have been populated so you won't need a password.  If you don't have SSH keys then the password is "rancher".

### Building

Docker 1.5+ required.

```
$ ./build.sh
```

When the build is done, the ISO should be in `dist/artifacts`.

### Repositories

All of repositories are located within our main GitHub [page](https://github.com/rancherio). 

[RancherOS Repo](https://github.com/rancherio/os): This repo contains the bulk of the RancherOS code.

[RancherOS Services Repo](https://github.com/rancherio/os-services): This repo is where any [system-services]({{site.baseurl}}/docs/system-services/) can be contributed.

[RancherOS Vagrant Repo]: We've created an easy way to spin up RancherOS using [Vagrant]({{site.baseurl}}/docs/getting-started/docs). 

If you have any updates to our documentation, please fork the `gh-pages` branch in our main [RancherOS Repo](https://github.com/rancherio/os) and contribute to that branch. 

### Bugs

If you find any bugs or are having any trouble, please contact us by filing an [issue](https://github.com/rancherio/os/issues/new). 

<br>
<br>