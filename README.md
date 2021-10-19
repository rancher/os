# Development and Maintenance Status
> RancherOS 1.x is no longer being actively maintained. There are two significant reasons behind this product decision:
1. **Docker** - The current industry requirements for a container runtime are very much evolving. Container runtimes like containerd and CRIO are now being actively considered as the default choice. RancherOS 1.x, which was specifically designed around using Docker engine only, unfortunately does not lend itself, in its current design, to this new evolving requirement.
2. **ISV Support** - RancherOS was specifically designed as a minimalistic OS to support purpose-built containerized applications. It was not designed to be used as a general purpose OS (such as CentOS or Ubuntu). As such, most ISVs have not certified their software to run on RancherOS, nor does RancherOS even contain the necessary components for many of these applications to run.

We're working on a replacement. Stay tuned!

# RancherOS

[![Build Status](https://drone-pr.rancher.io/api/badges/rancher/os/status.svg?branch=master)](https://drone-pr.rancher.io/rancher/os)
[![Docker Pulls](https://img.shields.io/docker/pulls/rancher/os.svg)](https://store.docker.com/community/images/rancher/os)
[![Go Report Card](https://goreportcard.com/badge/github.com/rancher/os)](https://goreportcard.com/badge/github.com/rancher/os)

The smallest, easiest way to run Docker in production at scale.  Everything in RancherOS is a container managed by Docker.  This includes system services such as udev and rsyslog.  RancherOS includes only the bare minimum amount of software needed to run Docker.  This keeps the binary download of RancherOS very small.  Everything else can be pulled in dynamically through Docker.

## How this works

Everything in RancherOS is a Docker container.  We accomplish this by launching two instances of
Docker.  One is what we call the system Docker which runs as the first process.  System Docker then launches
a container that runs the user Docker.  The user Docker is then the instance that gets primarily
used to create containers.  We created this separation because it seemed logical and also
it would really be bad if somebody did `docker rm -f $(docker ps -qa)` and deleted the entire OS.

![How it works](./rancheros.png "How it works")

## Release

- **v1.5.8 - Docker 19.03.15 - Linux 4.14.138**

### ISO

- https://releases.rancher.com/os/v1.5.8/rancheros.iso
- https://releases.rancher.com/os/v1.5.8/hyperv/rancheros.iso
- https://releases.rancher.com/os/v1.5.8/4glte/rancheros.iso
- https://releases.rancher.com/os/v1.5.8/vmware/rancheros.iso

#### Special docker-machine Links

- https://releases.rancher.com/os/v1.5.8/vmware/rancheros-autoformat.iso
- https://releases.rancher.com/os/v1.5.8/proxmoxve/rancheros-autoformat.iso

### Additional Downloads

#### AMD64 Links

* https://releases.rancher.com/os/v1.5.8/initrd
* https://releases.rancher.com/os/v1.5.8/vmlinuz
* https://releases.rancher.com/os/v1.5.8/rancheros.ipxe
* https://releases.rancher.com/os/v1.5.8/rootfs.tar.gz

#### ARM64 Links

* https://releases.rancher.com/os/v1.5.8/arm64/initrd
* https://releases.rancher.com/os/v1.5.8/arm64/vmlinuz
* https://releases.rancher.com/os/v1.5.8/arm64/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v1.5.8/arm64/rancheros-raspberry-pi64.zip

#### Cloud Links

* https://releases.rancher.com/os/v1.5.8/rancheros-openstack.img
* https://releases.rancher.com/os/v1.5.8/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.5.8/rancheros-cloudstack.img
* https://releases.rancher.com/os/v1.5.8/rancheros-aliyun.vhd
* https://releases.rancher.com/os/v1.5.8/rancheros-gce.tar.gz

#### VMware Links

* https://releases.rancher.com/os/v1.5.8/vmware/initrd
* https://releases.rancher.com/os/v1.5.8/vmware/rancheros.vmdk
* https://releases.rancher.com/os/v1.5.8/vmware/rootfs.tar.gz

#### Hyper-V Links

* https://releases.rancher.com/os/v1.5.8/hyperv/initrd
* https://releases.rancher.com/os/v1.5.8/hyperv/rootfs.tar.gz

#### Proxmox VE Links

* https://releases.rancher.com/os/v1.5.8/proxmoxve/initrd
* https://releases.rancher.com/os/v1.5.8/proxmoxve/rootfs.tar.gz

#### 4G-LTE Links

* https://releases.rancher.com/os/v1.5.8/4glte/initrd
* https://releases.rancher.com/os/v1.5.8/4glte/rootfs.tar.gz

**Note**:
1. you can use `http` instead of `https` in the above URLs, e.g. for iPXE.
2. you can use `latest` instead of `v1.5.8` in the above URLs if you want to get the latest version.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI
-------|------|------
eu-north-1 | HVM | [ami-08b189555c5d2d8c3](https://eu-north-1.console.aws.amazon.com/ec2/home?region=eu-north-1#launchInstanceWizard:ami=ami-08b189555c5d2d8c3)
ap-south-1 | HVM | [ami-0086964cb1ffc4fdb](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-0086964cb1ffc4fdb)
eu-west-3 | HVM | [ami-088930cafc1ad9f20](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-088930cafc1ad9f20)
eu-west-2 | HVM | [ami-0fdf07cfd187af004](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-0fdf07cfd187af004)
eu-west-1 | HVM | [ami-0cea454c576ececbd](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-0cea454c576ececbd)
ap-northeast-2 | HVM | [ami-0fdb6555f88256d12](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-0fdb6555f88256d12)
ap-northeast-1 | HVM | [ami-052c75c3e8757bcd9](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-052c75c3e8757bcd9)
sa-east-1 | HVM | [ami-04e51c9d1edad1bfd](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-04e51c9d1edad1bfd)
ca-central-1 | HVM | [ami-006a1ff3bf21003b3](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-006a1ff3bf21003b3)
ap-southeast-1 | HVM | [ami-03b14c67c74644c2b](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-03b14c67c74644c2b)
ap-southeast-2 | HVM | [ami-07059c8f12411bfcb](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-07059c8f12411bfcb)
eu-central-1 | HVM | [ami-0fc1a9332c246bc56](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-0fc1a9332c246bc56)
us-east-1 | HVM | [ami-02fe87f853d560d52](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-02fe87f853d560d52)
us-east-2 | HVM | [ami-004259f4c48585992](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-004259f4c48585992)
us-west-1 | HVM | [ami-0b382b76fadc95544](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-0b382b76fadc95544)
us-west-2 | HVM | [ami-0cdefa6a0646eb511](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-0cdefa6a0646eb511)

Additionally, images are available with support for Amazon EC2 Container Service (ECS) [here](https://rancher.com/docs/os/v1.x/en/installation/amazon-ecs/#amazon-ecs-enabled-amis).

### Azure

You can get RancherOS in the [Azure Marketplace](https://azuremarketplace.microsoft.com/en-us/marketplace/apps/rancher.rancheros), currently only the `rancher` user can be logged in through SSH keys.

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](https://rancher.com/docs/os/v1.x/en/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

## Support, Discussion, and Community
If you need any help with RancherOS or Rancher, please join us at either our [Rancher forums](http://forums.rancher.com) or [#rancher IRC channel](http://webchat.freenode.net/?channels=rancher) where most of our team hangs out at.

For security issues, please email security@rancher.com instead of posting a public issue in GitHub.  You may (but are not required to) use the GPG key located on [Keybase](https://keybase.io/rancher).


Please submit any **RancherOS** bugs, issues, and feature requests to [rancher/os](//github.com/rancher/os/issues).

Please submit any **Rancher** bugs, issues, and feature requests to [rancher/rancher](//github.com/rancher/rancher/issues).

## License

Copyright (c) 2014-2020 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
