---
title: Custom Kernels in RancherOS


---

## Custom Kernels
---

### Changing the Kernel in RancherOS

By default, RancherOS ships with the kernel provided by the [os-kernel repository](https://github.com/rancher/os-kernel). Swapping out the default kernel can by done by [building your own custom RancherOS ISO]({{page.osbaseurl}}/configuration/custom-rancheros-iso/).

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

 3. After you've replaced the URL with your custom kernel, you can follow the steps in [building your own custom RancherOS ISO]({{page.osbaseurl}}/configuration/custom-rancheros-iso/).

### Packaging a Kernel to be used in RancherOS

We build the kernel for RancherOS at the [os-kernel repository](https://github.com/rancher/os-kernel). You can use this repository to help package your own custom kernel to be used in RancherOS.


1. Create a clone of the [os-kernel](https://github.com/rancher/os-kernel) repository to your local machine using `git clone`.

   ```
   $ git clone https://github.com/rancher/os-kernel.git
   ```

2. In the `./scripts/build-common` file, update the `KERNEL_URL` and `KERNEL_SHA1`. `KERNEL_URL` points to Linux kernel sources archive, packaged as `.tar.gz` or `.tar.xz`. `KERNEL_SHA1` is the `SHA1` sum of the kernel sources archive.

   `./scripts/build-common` file

   ```bash
   #!/bin/bash
   set -e

   : ${KERNEL_URL:="https://github.com/rancher/linux/archive/Ubuntu-3.19.0-27.29.tar.gz"}
   : ${KERNEL_SHA1:="84b9bc53bbb4dd465b97ea54a71a9805e27ae4f2"}
   : ${ARTIFACTS:=$(pwd)/assets}
   : ${BUILD:=$(pwd)/build}
   : ${CONFIG:=$(pwd)/config}
   : ${DIST:=$(pwd)/dist}
   ```

3. After you've replaced the `KERNEL_URL` and `KERNEL_SHA1`, run `make` in the root `os-kernel` directory. After the build is completed, a `./dist/kernel` directory will be created with the freshly built kernel tarball and headers.

   ```
   $ make
   $ cd dist/kernel
   $ ls
   build.tar.gz                     extra.tar.gz    <name_of_kernel>.tar.gz
   ```

The `build.tar.gz` and `extra.tar.gz` files are used to build the `rancher/os-extras` and `rancher/os-headers` images for your RancherOS release - see https://github.com/rancher/os-images and https://github.com/rancher/os-services for how to build the images and make them available.
