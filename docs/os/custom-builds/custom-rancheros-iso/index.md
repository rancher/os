---
title: Custom RancherOS ISO
layout: os-default
redirect_from:
  - os/configuration/custom-rancheros-iso/
---

## Custom RancherOS ISO

It's easy to build your own RancherOS ISO.

1. Create a clone of the main [RancherOS repository](https://github.com/rancher/os) to your local machine with a `git clone`.

   ```
   $ git clone https://github.com/rancher/os.git
   ```

2. In the root of the repository, the "General Configuration" section of `Dockerfile.dapper` can be updated to use [custom kernels]({{site.baseurl}}/os/configuration/custom-kernels), or [custom Docker]({{site.baseurl}}/os/configuration/custom-docker).

3. After you've saved your edits, run `make` in the root directory. After the build has completed, a `./dist/artifacts` directory will be created with the custom built RancherOS release files.

     Build Requirements: `bash`, `make`, `docker` (Docker version >= 1.10.3)

   ```
   $ make
   $ cd dist/artifacts
   $ ls
   initrd             rancheros.iso
   iso-checksums.txt	vmlinuz
   ```

The `rancheros.iso` is ready to be used to [boot RancherOS from ISO]({{site.baseurl}}/os/running-rancheros/workstation/boot-from-iso/) or [launch RancherOS using Docker Machine]({{site.baseurl}}/os/running-rancheros/workstation/docker-machine). 


### Creating a GCE Image Archive

You can build the [GCE image archive](https://cloud.google.com/compute/docs/tutorials/building-images) using [Packer](https://www.packer.io/). You will need Packer, QEMU and GNU tar installed.

First, create `gce-qemu.json`:

```json
{
 "builders":
 [
   {
     "type": "qemu",
     "name": "qemu-googlecompute",
     "iso_url": "https://github.com/rancherio/os/releases/download/<RancherOS-Version>/rancheros.iso",
     "iso_checksum": "<rancheros.iso-MD5-hash>",
     "iso_checksum_type": "md5",
     "ssh_wait_timeout": "360s",
     "disk_size": 10000,
     "format": "raw",
     "headless": true,
     "accelerator": "none",
     "ssh_host_port_min": 2225,
     "ssh_host_port_max": 2229,
     "ssh_username": "rancher",
     "ssh_password": "rancher",
     "ssh_port": 22,
     "net_device": "virtio-net",
     "disk_interface": "scsi",
     "qemuargs": [
       ["-m", "1024M"], ["-nographic"], ["-display", "none"]
     ]
   }
 ],
 "provisioners": [
   {
     "type":"shell",
     "script": "../scripts/install2disk"
   }
 ]
}
```

NOTE: For faster builds You can use `"kvm"` as the `accelerator` field value if you have KVM, but that's optional.

Run:

```
$ packer build gce-qemu.json
```

Packer places its output into `output-qemu-googlecompute/packer-qemu-googlecompute` - it's a raw VM disk image. Now you just need to name it `disk.raw` and package it as sparse .tar.gz:

```
$ mv output-qemu-googlecompute/packer-qemu-googlecompute disk.raw
$ tar -czSf rancheros-<RancherOS-Version>.tar.gz disk.raw
```

NOTE: the last command should be using GNU tar. It might be named `gtar` on your system.
