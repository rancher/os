---
title: Built-in System Services in RancherOS


---

## Built-in System Services
---

To launch RancherOS, we have built-in system services. They are defined in the [Docker Compose](https://docs.docker.com/compose/compose-file/) format, and can be found in the default system config file, `/usr/share/ros/os-config.yml`. You can [add your own system services]({{page.osbaseurl}}/system-services/) or override services in the cloud-config.

In start up order, here are the groups of services:

1. Device and power management:
- udev-cold
- udev
- acpid

2. syslog

3. System configuration and networking:
- preload-system-images
- cloud-init-pre
- network-pre
- ntp
- cloud-init
- network

4. User interaction:
- console
- docker

5. Post configuration:
- preload-user-images

### preload-system-images & preload-user-images

Read more about [pre-packing Docker images]({{page.osbaseurl}}/configuration/prepacking-docker-images/).

### cloud-init-pre

User-data (i.e. [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config)) and metadata from cloud provider, VM runtime, or a management service, is loaded in this service.

The user-data is written to:

* `/var/lib/rancher/conf/cloud-config.d/boot.yml` - If the user-data is a cloud-config, i.e. begins with `#cloud-config` and is YAML format.
* `/var/lib/rancher/conf/cloud-config-script` - If the user-data is a script, i.e begins with `#!`.
* `/var/lib/rancher/conf/metadata` - If it is serialized cloud provider metadata.

It is configured by the `rancher.cloud_init.datasources` list in [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config). It is pre-configured in cloud-provider specific images (e.g. AWS, GCE).

### network-pre

During this service, networking is set up, e.g. hostname, interfaces, and DNS.

It is configured by `hostname` and `rancher.network`[settings]({{page.osbaseurl}}/networking/) in [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config).

### ntp

Runs `ntpd` in a System Docker container.

### cloud-init

It does the same thing as cloud-init-pre, but in this step, it can also use the network to fetch user-data and metadata (e.g. in cloud providers).


### network

Completes setting up networking with configuration obtained by cloud-init.


### console

This service provides the RancherOS user interface by running `sshd` and `getty`. It completes the RancherOS configuration on start up:

1. If the `rancher.password=<password>` kernel parameter exists, it sets `<password>` as the password for the `rancher` user.

2. If there are no host SSH keys, it generates host SSH keys and saves them under `rancher.ssh.keys` in [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config).

3. Runs `cloud-init -execute`, which does the following:

   * Updates `.ssh/authorized_keys` in `/home/rancher` and `/home/docker` from [cloud-config]({{page.osbaseurl}}/configuration/ssh-keys/) and metadata.
   * Writes files specified by the `write_files` [cloud-config]({{page.osbaseurl}}/configuration/write-files/) setting.
   * Resizes the device specified by the `rancher.resize_device` [cloud-config]({{page.osbaseurl}}/configuration/resizing-device-partition/) setting.
   * Mount devices specified in the `mounts` [cloud-config]({{page.osbaseurl}}/configuration/additional-mounts/) setting.
   * Set sysctl parameters specified in  the`rancher.sysctl` [cloud-config]({{page.osbaseurl}}/configuration/sysctl/) setting.

4. If user-data contained a file that started with `#!`, then a file would be saved at `/var/lib/rancher/conf/cloud-config-script` during cloud-init and then executed. Any errors are ignored.

5. Runs `/opt/rancher/bin/start.sh` if it exists and is executable. Any errors are ignored.

6. Runs `/etc/rc.local` if it exists and is executable. Any errors are ignored.


### docker

This system service runs the user docker daemon. Normally it runs inside the console system container by running `docker-init` script which, in turn, looks for docker binaries in `/opt/bin`, `/usr/local/bin` and `/usr/bin`, adds the first found directory with docker binaries to PATH and runs `dockerlaunch docker daemon` appending the passed arguments.

Docker daemon args are read from `rancher.docker.args` cloud-config property (followed by `rancher.docker.extra_args`).

### RancherOS Configuration Load Order

[Cloud-config]({{page.osbaseurl}}/configuration/#cloud-config/) is read by system services when they need to get configuration. Each additional file overwrites and extends the previous configuration file.

1. `/usr/share/ros/os-config.yml` - This is the system default configuration, which should **not** be modified by users.
2. `/usr/share/ros/oem/oem-config.yml` - This will typically exist by OEM, which should **not** be modified by users.
3. Files in `/var/lib/rancher/conf/cloud-config.d/` ordered by filename. If a file is passed in through user-data, it is written by cloud-init and saved as `/var/lib/rancher/conf/cloud-config.d/boot.yml`.
4. `/var/lib/rancher/conf/cloud-config.yml` - If you set anything with `ros config set`, the changes are saved in this file.
5. Kernel parameters with names starting with `rancher`.
6. `/var/lib/rancher/conf/metadata` - Metadata added by cloud-init.
