# Distribution via Config Drive

CoreOS supports providing configuration data via [config drive][config-drive]
disk images. Currently only providing a single script or cloud config file is
supported.

[config-drive]: http://docs.openstack.org/user-guide/cli_config_drive.html

## Contents and Format

The image should be a single FAT or ISO9660 file system with the label
`config-2` and the configuration data should be located at
`openstack/latest/user_data`.

For example, to wrap up a config named `user_data` in a config drive image:

```sh
mkdir -p /tmp/new-drive/openstack/latest
cp user_data /tmp/new-drive/openstack/latest/user_data
mkisofs -R -V config-2 -o configdrive.iso /tmp/new-drive
rm -r /tmp/new-drive
```

If on OS X, replace the `mkisofs` invocation with:

```sh
hdiutil makehybrid -iso -joliet -default-volume-name config-2 -o configdrive.iso /tmp/new-drive
```

## QEMU virtfs

One exception to the above, when using QEMU it is possible to skip creating an
image and use a plain directory containing the same contents:

```sh
qemu-system-x86_64 \
    -fsdev local,id=conf,security_model=none,readonly,path=/tmp/new-drive \
    -device virtio-9p-pci,fsdev=conf,mount_tag=config-2 \
    [usual qemu options here...]
```
