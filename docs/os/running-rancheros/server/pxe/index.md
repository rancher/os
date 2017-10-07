---
title: Booting RancherOS with iPXE


---
## Booting RancherOS via iPXE
----

```
#!ipxe
# Boot a persistent RancherOS to RAM

# Location of Kernel/Initrd images
set base-url http://releases.rancher.com/os/latest

kernel ${base-url}/vmlinuz rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda] rancher.cloud_init.datasources=[url:http://example.com/cloud-config]
initrd ${base-url}/initrd
boot
```

### Hiding sensitive kernel commandline parameters

From RancherOS v0.9.0, secrets can be put on the `kernel` parameters line afer a `--` double dash, and they will be not be shown in any `/proc/cmdline`. These parameters
will be passed to the RancherOS init process and stored in the `root` accessible `/var/lib/rancher/conf/cloud-init.d/init.yml` file, and are available to the root user from the `ros config` commands.

For example, the `kernel` line above could be written as:

```
kernel ${base-url}/vmlinuz rancher.state.dev=LABEL=RANCHER_STATE rancher.state.autoformat=[/dev/sda] -- rancher.cloud_init.datasources=[url:http://example.com/cloud-config]
```

The hidden part of the command line can be accessed with either `sudo ros config get rancher.environment.EXTRA_CMDLINE`, or by using a service file's environment array.

An example service.yml file:

```
test:
  image: alpine
  command: echo "tell me a secret ${EXTRA_CMDLINE}"
  labels:
    io.rancher.os.scope: system
  environment:
  - EXTRA_CMDLINE
```

When this service is run, the `EXTRA_CMDLINE` will be set.


### cloud-init Datasources

Valid cloud-init datasources for RancherOS.

| type | default |  |
|---|---|--|
| ec2 | ec2's DefaultAddress |  |
| file | path |  |
| cmdline | /media/config-2 |  |
| configdrive |  |  |
| digitalocean | DefaultAddress |  |
| ec2 | DefaultAddress |  |
| file | path |  |
| gce |  |  |
| packet | DefaultAddress |  |
| openstack| OpenStack's DefaultAddress |  |
| url | url |  |
| vmware |  | set `guestinfo` cloud-init or interface data as per [VMware ESXi]({{page.osbaseurl}}/cloud/vmware-esxi) |
| * | This will add ["configdrive", "vmware", "ec2", "digitalocean", "packet", "gce"] into the list of datasources to try |  |

### Cloud-Config

When booting via iPXE, RancherOS can be configured using a [cloud-config file]({{page.osbaseurl}}/configuration/#cloud-config).

