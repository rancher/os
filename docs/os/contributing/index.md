---
title: Contributing to RancherOS


---

## Contributing to RancherOS
---

## Developing

Development is easiest done with QEMU on Linux. OS X works too, although QEMU doesn't have KVM support. If you are running Linux in a virtual machine, then we recommend you run VMWare Fusion/Workstation and enable VT-x support.  Then, QEMU will have KVM support and run sufficiently fast inside your Linux VM.

### Building

#### Requirements:

* bash
* make
* Docker 1.10.3+

```
$ make
```

The build will run in Docker containers, and when the build is done, the vmlinuz, initrd, and ISO should be in `dist/artifacts`.

If you're building a version of RancherOS used for development and not for a release, you can instead run `make dev`. This will run faster than the standard build by avoiding building the `installer.tar` and `rootfs.tar.gz` artifacts which are not needed by QEMU.

### Testing

Run `make integration-tests` to run the all integration tests in a container, or `./scripts/integration-tests` to run them outside a container (they use QEMU to test the OS.)

To run just one integration test, or a group of them (using regex's like `.*Console.*`, you can set the `RUNTEST` environment variable:

```
$ RUNTEST=TestPreload make integration-test
```

### Running

Prerequisites: QEMU, coreutils, cdrtools/genisoimage/mkisofs.
On OS X, `brew` is recommended to install those. On Linux, use your distro package manager.

To launch RancherOS in QEMU from your dev version, you can either use `make run`, or customise the vm using `./scripts/run` and its options. You can use `--append your.kernel=params here` and `--cloud-config your-cloud-config.yml` to configure the RancherOS instance you're launching.

You can SSH in using `./scripts/ssh`.  Your SSH keys should have been populated (if you didn't provide your own cloud-config) so you won't need a password.  If you don't have SSH keys, or something is wrong with your cloud-config, then the password is "`rancher`".

If you're on OS X, you can run RancherOS using [_xhyve_](https://github.com/mist64/xhyve) instead of QEMU: just pass `--xhyve` to `./scripts/run` and `./scripts/ssh`.

### Debugging and logging.

You can enable extra log information in the console by setting them using `sudo ros config set`,
or as kernel boot parameters.
Enable all logging by setting `rancher.debug` true
or you can set `rancher.docker.debug`, `rancher.system_docker.debug`, `rancher.bootstrap_docker.debug`, or `rancher.log` individually.

You will also be able to view the debug logging information by running `dmesg` as root.

## Repositories

All of repositories are located within our main GitHub [page](https://github.com/rancher).

[RancherOS Repo](https://github.com/rancher/os): This repo contains the bulk of the RancherOS code.

[RancherOS Services Repo](https://github.com/rancher/os-services): This repo is where any [system-services]({{page.osbaseurl}}/system-services/) can be contributed.

[RancherOS Images Repo](https://github.com/rancher/os-images): This repo is for the corresponding service images.


## Bugs

If you find any bugs or are having any trouble, please contact us by filing an [issue](https://github.com/rancher/os/issues/new).

If you have any updates to our documentation, please make any PRs to our [docs repo](https://github.com/rancher/rancher.github.io).

<br>
<br>
