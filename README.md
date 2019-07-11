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

- **v1.5.3 - Docker 18.06.3-ce - Linux 4.14.128**

### ISO

- https://releases.rancher.com/os/v1.5.3/rancheros.iso
- https://releases.rancher.com/os/v1.5.3/hyperv/rancheros.iso
- https://releases.rancher.com/os/v1.5.3/4glte/rancheros.iso
- https://releases.rancher.com/os/v1.5.3/vmware/rancheros.iso

#### Special docker-machine Links

- https://releases.rancher.com/os/v1.5.3/vmware/rancheros-autoformat.iso
- https://releases.rancher.com/os/v1.5.3/proxmoxve/rancheros-autoformat.iso

### Additional Downloads

#### AMD64 Links

* https://releases.rancher.com/os/v1.5.3/initrd
* https://releases.rancher.com/os/v1.5.3/vmlinuz
* https://releases.rancher.com/os/v1.5.3/rancheros.ipxe
* https://releases.rancher.com/os/v1.5.3/rootfs.tar.gz

#### ARM64 Links

* https://releases.rancher.com/os/v1.5.3/arm64/initrd
* https://releases.rancher.com/os/v1.5.3/arm64/vmlinuz
* https://releases.rancher.com/os/v1.5.3/arm64/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v1.5.3/arm64/rancheros-raspberry-pi64.zip

#### Cloud Links

* https://releases.rancher.com/os/v1.5.3/rancheros-openstack.img
* https://releases.rancher.com/os/v1.5.3/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.5.3/rancheros-cloudstack.img
* https://releases.rancher.com/os/v1.5.3/rancheros-aliyun.vhd
* https://releases.rancher.com/os/v1.5.3/rancheros-gce.tar.gz

#### VMware Links

* https://releases.rancher.com/os/v1.5.3/vmware/initrd
* https://releases.rancher.com/os/v1.5.3/vmware/rancheros.vmdk
* https://releases.rancher.com/os/v1.5.3/vmware/rootfs.tar.gz

#### Hyper-V Links

* https://releases.rancher.com/os/v1.5.3/hyperv/initrd
* https://releases.rancher.com/os/v1.5.3/hyperv/rootfs.tar.gz

#### Proxmox VE Links

* https://releases.rancher.com/os/v1.5.3/proxmoxve/initrd
* https://releases.rancher.com/os/v1.5.3/proxmoxve/rootfs.tar.gz

#### 4G-LTE Links

* https://releases.rancher.com/os/v1.5.3/4glte/initrd
* https://releases.rancher.com/os/v1.5.3/4glte/rootfs.tar.gz

**Note**:
1. you can use `http` instead of `https` in the above URLs, e.g. for iPXE.
2. you can use `latest` instead of `v1.5.3` in the above URLs if you want to get the latest version.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI
-------|------|------
eu-north-1 | HVM | [ami-0d0ee5e9b88ab3641](https://eu-north-1.console.aws.amazon.com/ec2/home?region=eu-north-1#launchInstanceWizard:ami=ami-0d0ee5e9b88ab3641)
ap-south-1 | HVM | [ami-0e64d18d3546bc0e3](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-0e64d18d3546bc0e3)
eu-west-3 | HVM | [ami-0f1cfb7093ca6f8c6](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-0f1cfb7093ca6f8c6)
eu-west-2 | HVM | [ami-0fb2092ee913a1deb](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-0fb2092ee913a1deb)
eu-west-1 | HVM | [ami-01bcb29c6aebb24bb](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-01bcb29c6aebb24bb)
ap-northeast-2 | HVM | [ami-0c52edead8ffab5d0](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-0c52edead8ffab5d0)
ap-northeast-1 | HVM | [ami-0b6f89180ebf76c13](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-0b6f89180ebf76c13)
sa-east-1 | HVM | [ami-04ab75c51a5fc9be6](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-04ab75c51a5fc9be6)
ca-central-1 | HVM | [ami-0dd7a2d93147d7eff](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-0dd7a2d93147d7eff)
ap-southeast-1 | HVM | [ami-058ff3e97841ed000](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-058ff3e97841ed000)
ap-southeast-2 | HVM | [ami-07bc506efd2ade5f1](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-07bc506efd2ade5f1)
eu-central-1 | HVM | [ami-0143355d2c9ff8db5](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-0143355d2c9ff8db5)
us-east-1 | HVM | [ami-0a48e384fc0262754](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-0a48e384fc0262754)
us-east-2 | HVM | [ami-03cc2d26e5ad831ea](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-03cc2d26e5ad831ea)
us-west-1 | HVM | [ami-0d1728efec5c95f06](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-0d1728efec5c95f06)
us-west-2 | HVM | [ami-0d5f95b9a27dfef6f](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-0d5f95b9a27dfef6f)
cn-north-1 | HVM | [ami-08e1d5d0aa44a7606](https://cn-north-1.console.amazonaws.cn/ec2/home?region=cn-north-1#launchInstanceWizard:ami=ami-08e1d5d0aa44a7606)
cn-northwest-1 | HVM | [ami-0847a5ca9e1b24751](https://cn-northwest-1.console.amazonaws.cn/ec2/home?region=cn-northwest-1#launchInstanceWizard:ami=ami-0847a5ca9e1b24751)

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

Copyright (c) 2014-2019 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
