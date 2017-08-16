---
title: DKMS / Loadable Kernel Modules in RancherOS


---

## Dynamic Kernel Module Support (DKMS) / Loadable Kernel Modules (LKM)

To compile any Kernel Modules, you first need to [deploy the Kernel Headers]({{site.baseurl}}/os/configuration/kernel-modules-kernel-headers/).

### DKMS

DKMS is supported by running the DKMS scripts inside a *privileged* container.

> To deploy containers that compiles DKMS modules, you will need to ensure that you bind-mount `/usr/src` and `/lib/modules`.

> To deploy containers that run any DKMS operations (i.e., `modprobe`), you will need to ensure that you bind-mount `/lib/modules`.

By default, the `/lib/modules` folder is already available in the console deployed via [RancherOS System Services]({{site.baseurl}}/os/system-services/built-in-system-services/), but not `/usr/src`. You will likely need to [deploy your own container](#docker-example) for compilation purposes.

To learn more about Docker's privileged mode, or to limit capabilities, please review the [Docker Runtime privilege and Linux capabilities documentation](https://docs.docker.com/engine/reference/run/#/runtime-privilege-and-linux-capabilities).

#### cloud-config Example

```yaml
myservice:
  image: ...
  privileged: true
  volumes:
  - /lib/modules:/lib/modules
  - /usr/src:/usr/src
```

#### Docker Example

> For one-off operations, it's useful to use `--rm` to clean up containers when operations complete.

```bash
$ sudo system-docker run -it --rm --name dkms-install -v /usr/src:/usr/src -v /lib/modules:/lib/modules ubuntu sh -c 'apt-get update && apt-get install -y sysdig-dkms'
```

The same approach can be utilized with the User Docker Daemon, just replace `sudo system-docker` with `docker`.

### LKM Dependencies

In some situations, another Kernel Module might need loading prior to any module you're trying to add.

In this example, we'll reference the `v4l2loopback` DKMS module, which requires probing `videodev` into the Kernel space and is not on any filesystem by default.

First, you must enable `kernel-extras`, then `modprobe` your dependencies and subsequent modules:

```bash
sudo ros service enable kernel-extras
sudo ros service up -d kernel-extras
```

This will overlay all the compiled modules into `/lib/modules/$(uname -r)` that are configured in the default RancherOS Kernel config.

Now you are ready to add your Modules into the Kernel space:

```bash
sudo modprobe videodev
sudo modprobe v4l2loopback
```

To see which modules are pre-built, you can either do a listing of all `.ko` (kernel object) files, or review the Kernel config:

```bash
find /lib*/modules/$(uname -r) -name *.ko | less
#or
zcat /proc/config.gz | less
```

For more information regarding modifying the Kernel, please review the [Custom Kernels]({{site.baseurl}}/os/custom-builds/custom-kernels/) documentation.

### Auto-Loading Modules

Kernel Modules can be automatically loaded with the `rancher.modules` cloud-config field.

```yaml
#cloud-config
rancher:
  modules: [btrfs]
```

This functionality is also available via a kernel parameter. For example, the btrfs module could be automatically loaded with `rancher.modules=[btrfs]` as a kernel parameter.

### Ubuntu-based Kernel Manipulation

For images that are or derive from Ubuntu, you will need some small packages for `depmod`(`kmod`) and `modprobe`(`module-init-tools`):

```bash
sudo apt-get install kmod module-init-tools
```

Most packages should already list these as dependencies in Aptitude, as well as `gcc` and related libs for packages that require compilation (which is most).

### Troubleshooting

Messing around with the Kernel can be tricky, so here's some common issues:

#### kernel source for this kernel does not seem to be installed.

Simply put, the Kernel Headers (or Source) cannot be found; enable them via the [Kernel Headers System Service]({{site.baseurl}}/os/configuration/kernel-modules-kernel-headers/).

#### Operation not Permitted

When inside a container, you might see similar to the following:
```
modprobe: ERROR: could not insert 'videodev': Operation not permitted
```

This is in reference to your container's privileges, not your user (i.e., `sudo` will not fix this).

Instead, ensure you started the container with `--privileged` or the `cloud-config` setting described above.

#### modprobe: ERROR: could not insert 'v4l2loopback': Unknown symbol in module, or unknown parameter (see dmesg)

Again, using `v4l2loopback` as an example, but this can happen for any module.

As stated, check out `dmesg` to see what the issue is. Chances are you'll see something like the following:

```bash
[  322.734052] v4l2loopback: module verification failed: signature and/or required key missing - tainting kernel
[  322.734141] v4l2loopback: Unknown symbol video_ioctl2 (err 0)
[  322.734454] v4l2loopback: Unknown symbol v4l2_ctrl_handler_init_class (err 0)
[  322.734526] v4l2loopback: Unknown symbol video_devdata (err 0)
[  322.734563] v4l2loopback: Unknown symbol v4l2_ctrl_new_custom (err 0)
[  322.734599] v4l2loopback: Unknown symbol video_unregister_device (err 0)
[  322.734635] v4l2loopback: Unknown symbol video_device_alloc (err 0)
[  322.734696] v4l2loopback: Unknown symbol v4l2_device_register (err 0)
[  322.734732] v4l2loopback: Unknown symbol __video_register_device (err 0)
[  322.734765] v4l2loopback: Unknown symbol v4l2_ctrl_handler_free (err 0)
[  322.734796] v4l2loopback: Unknown symbol v4l2_device_unregister (err 0)
[  322.734828] v4l2loopback: Unknown symbol video_device_release (err 0)
```

This one can be trickier to evaluate, so start searching Google for symbol names to figure out which modules they derive from.

In this example, `video_ioctl2` comes from `videodev` and can be simply inserted via the `kernel-extras` overlay described above.
