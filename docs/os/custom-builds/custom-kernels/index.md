---
title: Custom Kernels in RancherOS
layout: os-default
redirect_from:
  - os/configuration/custom-kernels/
---

## Custom Kernels

### Building and Packaging a Kernel to be used in RancherOS

We build the kernel for RancherOS at the [os-kernel repository](https://github.com/rancher/os-kernel). You can use this repository to help package your own custom kernel to be used in RancherOS.


1. Create a clone of the [os-kernel](https://github.com/rancher/os-kernel) repository to your local machine using `git clone`.
    
   ```
   $ git clone https://github.com/rancher/os-kernel.git
   ```

2. In the `./Dockerfile.dapper` file, update the `KERNEL_TAG`, `KERNEL_VERSION`, `KERNEL_URL` and `KERNEL_SHA1`. `KERNEL_URL` points to Linux kernel sources archive, packaged as `.tar.gz` or `.tar.xz`. `KERNEL_SHA1` is the `SHA1` sum of the kernel sources archive. 

   `./Dockerfile.dapper` file

   ```bash
   ########## Kernel version Configuration #############################
   ENV KERNEL_TAG=v4.8.7
   ENV KERNEL_VERSION=4.8.7-rancher
   ENV KERNEL_SHA1=5c10724a0e7e97b72046be841df0c69c6e2a03c2
   ENV KERNEL_URL=https://github.com/rancher/linux/archive/${KERNEL_TAG}.tar.gz
   ```

3. After you've replaced the `KERNEL_*` values, run `make` in the root `os-kernel` directory. After the build is completed, a `./dist/kernel` directory will be created with the freshly built kernel tarball and headers. 
   
   ```
   $ make
   ...snip...
   --- 4.8.7-rancher Kernel prepared for RancherOS
   	./dist/kernel/extra-linux-4.8.7-rancher-x86.tar.gz
   	./dist/kernel/build-linux-4.8.7-rancher-x86.tar.gz
   	./dist/kernel/linux-4.8.7-rancher-x86.tar.gz
   	./dist/kernel/config
   
   Images ready to push:
   rancher/os-extras:4.8.7-rancher
   rancher/os-headers:4.8.7-rancher

   ```

Now you need to either upload the `./dist/kernel/linux-4.8.7-rancher-x86.tar.gz` file to somewhere, or copy that file into your clone of the `rancher/os` repo, as `assets/kernel.tar.gz`.

The `build-<name>.tar.gz` and `extra-<name>.tar.gz` files are used to build the `rancher/os-extras` and `rancher/os-headers` images for your RancherOS release - which you will need to tag them with a different organisation name, push them to a registry, and create custom service.yml files.

### Building a RancherOS release using the Packaged kernel files.

By default, RancherOS ships with the kernel provided by the [os-kernel repository](https://github.com/rancher/os-kernel). Swapping out the default kernel can by done by [building your own custom RancherOS ISO]({{site.baseurl}}/os/configuration/custom-rancheros-iso/).

 1. Create a clone of the main [RancherOS repository](https://github.com/rancher/os) to your local machine with a `git clone`. 

    ```
    $ git clone https://github.com/rancher/os.git
    ```

 2. In the root of the repository, the "General Configuration" section of `Dockerfile.dapper` will need to be updated. Using your favorite editor, replace the appropriate `KERNEL_URL` value with a URL of your compiled custom kernel tarball. Ideally, the URL will use `HTTPS`.

    `Dockerfile.dapper` file

    ```
    # Update the URL to your own custom kernel tarball
    ARG KERNEL_URL_amd64=https://github.com/rancher/os-kernel/releases/download/Ubuntu-4.4.0-23.41-rancher/linux-4.4.10-rancher-x86.tar.gz
    ARG KERNEL_URL_arm64=https://github.com/imikushin/os-kernel/releases/download/Estuary-4.1.18-arm64-3/linux-4.1.18-arm64.tar.gz
    ```

    <br>

    > **Note:** `KERNEL_URL` settings should point to a Linux kernel, compiled and packaged in a specific way. You can fork [os-kernel repository](https://github.com/rancher/os-kernel) to package your own kernel.

    Your kernel should be packaged and published as a set of files of the following format:

    `<kernel-name-and-version>.tar.gz` is the one KERNEL_URL should point to. It contains the kernel binary, core modules and firmware:

    ```
    boot/
         vmlinuz-<kernel-version>
    lib/
        modules/
                <kernel-version>/
                                 ...
        firmware/
                 ...
    ```

    `build.tar.gz` contains build headers to build additional modules (e.g. using DKMS): it is a subset of the kernel sources tarball. These files will be installed into `/usr/src/<os-kernel-tag>` using the `kernel-headers-system-docker` and `kernel-headers` services.

    `extra.tar.gz` contains extra modules and firmware for your kernel and should be built into a `kernel-extras` service:

    ```
    lib/
        modules/
                <kernel-version>/
                                 ...
        firmware/
                 ...
    ```
  
 3. After you've replaced the URL with your custom kernel, you can follow the steps in [building your own custom RancherOS ISO]({{site.baseurl}}/os/configuration/custom-rancheros-iso/).
