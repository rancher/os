# Installation

## Building Images

The design of RancherOS is that you boot from a vanilla image and through cloud-init and Kubernetes mechanisms
the node will be configured. Installation of RancherOS is really the process of building an image from which
you can boot.  During the image building process you can bake in default OEM configuration that is a part of the
image.

## Installation Configuration

The installation process is driven by a config file. The configuration file contains the installation directives and
the OEM configuration for the image.

The installation configuration should be hosted on an HTTP or TFTP server. A simple approach is to use a
[GitHub Gist](https://gist.github.com).

### Kernel Command Line

Install directives can be set from the kernel command line using a period (.) seperated key structure such as
`rancheros.install.configurl`.  They kernel command line keys are case-insensitive.

### Reference

```yaml
#cloud-config
rancheros:
  install:
    # An http://, https://, or tftp:// URL to load as the base configuration
    # for this configuration. This configuration can include any install 
    # directives or OEM configuration. The resulting merged configuration
    # will be read by the installer and all content of the merged config will
    # be stored in /oem/99_custom.yaml in the created image.
    configURL: http://example.com/machine-cloud-config
    # Turn on verbose logging for the installation process
    debug: false
    # The target device that will be formatted and grub will be install on.
    # The partition table will be cleared and recreated with the default
    # partition layout. If noFormat is set to true this parameter is only
    # used to install grub.
    device: /dev/vda
    # If the system has the path /sys/firmware/efi it will be treated as a
    # UEFI system. If you are created an UEFI image on a non-EFI platform
    # then this flag will force the installer to use UEFI even if not detected.
    forceEFI: false
    # If true then it is assumed that the disk is already formatted with the standard
    # partitions need by RancherOS.  Refer to the partition table section below for the
    # exact requirements. Also, if this is set to true 
    noFormat: false
    # After installation the system will reboot by default.  If you wish to instead
    # power off the system set this to true.
    powerOff: false
    # The installed image will setup the default console to the current TTY value
    # used during the installation.  To force the installation to use a different TTY
    # then set that value here.
    tty: ttyS0

# Any other cloud-init values can be included in this file and will be stored in
# /oem/99_custom.yaml of the installed image
```

## ISO Installation

When booting from the ISO you will immediately be presented with the shell. The root password is hard coded to `root/ros`
if needed. An SSH server will be running so realize that __because of the hard coded password this is an insecure
system__ to be running on a public network.

From the shell run the below where `${LOCATION}` should be a path to a local file or `http://`, `https://`, or
`tftp://` URL.

```bash
ros-installer -config-file ${LOCATION}
```

### Interactive

`ros-installer` can also be ran without any arguments to allow you to install a simple vanilla image with a
root password set.

## iPXE Installation

Use the following script to install RancherOS using iPXE

```
#!ipxe
set config http://example.com/machine-cloud-confg
set version v0.0.1-test1
set arch amd64
set cmdline "sample.extra.cmdline.param=1"
chain https://github.com/rancher/os2/releases/download/${version}/rancheros-${version}-${arch}.ipxe
```

## Partition Table

RancherOS requires the following partitions.  These partitions are required by [cOS-toolkit](https://rancher-sandbox.github.io/cos-toolkit-docs/docs)

| Label          | Default Size    | Contains                                                    |
| ---------------|-----------------|------------------------------------------------------------ |
| COS_BOOT       |          50 MiB | UEFI Boot partition                                         |
| COS_STATE      |          15 GiB | A/B bootable file system images constructed from OCI images |
| COS_OEM        |          50 MiB | OEM cloud-config files and other data                       |
| COS_RECOVERY   |           8 GiB | Recovery file system image if COS_STATE is destroyed        |
| COS_PERSISTENT | Remaining space | All contents of the persistent folders                      |

## Folders

| Path              | Read-Only | Ephemeral | Persistent |
| ------------------|:---------:|:---------:|:----------:|
| /                 | x         |           |            |
| /etc              |           | x         |            |
| /etc/cni          |           |           | x          |
| /etc/iscsi        |           |           | x          |
| /etc/rancher      |           |           | x          |
| /etc/ssh          |           |           | x          |
| /etc/systemd      |           |           | x          |
| /srv              |           | x         |            |
| /home             |           |           | x          |
| /opt              |           |           | x          |
| /root             |           |           | x          |
| /var              |           | x         |            |
| /usr/libexec      |           |           | x          |
| /var/lib/cni      |           |           | x          |
| /var/lib/kubelet  |           |           | x          |
| /var/lib/longhorn |           |           | x          |
| /var/lib/rancher  |           |           | x          |
| /var/lib/wicked   |           |           | x          |
| /var/log          |           |           | x          |
