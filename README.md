# RancherOS

The smallest, easiest way to run Docker in production at scale.  Everything in RancherOS is a container managed by Docker.  This includes system services such as udev and rsyslog.  RancherOS includes only the bare minimum amount of software needed to run Docker.  This keeps the binary download of RancherOS very small.  Everything else can be pulled in dynamically through Docker.

## How this works

Everything in RancherOS is a Docker container.  We accomplish this by launching two instances of
Docker.  One is what we call the system Docker which runs as the first process.  System Docker then launches
a container that runs the user Docker.  The user Docker is then the instance that gets primarily
used to create containers.  We created this separation because it seemed logical and also
it would really be bad if somebody did `docker rm -f $(docker ps -qa)` and deleted the entire OS.

![How it works](./rancheros.png "How it works")

## Release

- **Latest: v1.4.2 - Docker 18.03.1-ce - Linux 4.14.73**
- **Stable: v1.4.2 - Docker 18.03.1-ce - Linux 4.14.73**

### ISO

- https://releases.rancher.com/os/latest/rancheros.iso
- https://releases.rancher.com/os/v1.4.2/rancheros.iso

### Additional Downloads

#### Latest Links

* https://releases.rancher.com/os/latest/initrd
* https://releases.rancher.com/os/latest/iso-checksums.txt
* https://releases.rancher.com/os/latest/rancheros-openstack.img
* https://releases.rancher.com/os/latest/rancheros-digitalocean.img
* https://releases.rancher.com/os/latest/rancheros-cloudstack.img
* https://releases.rancher.com/os/latest/rancheros-aliyun.vhd
* https://releases.rancher.com/os/latest/rancheros.ipxe
* https://releases.rancher.com/os/latest/rancheros-gce.tar.gz
* https://releases.rancher.com/os/latest/rootfs.tar.gz
* https://releases.rancher.com/os/latest/vmlinuz
* https://releases.rancher.com/os/latest/rancheros-vmware.iso

#### v1.4.2 Links

* https://releases.rancher.com/os/v1.4.2/initrd
* https://releases.rancher.com/os/v1.4.2/iso-checksums.txt
* https://releases.rancher.com/os/v1.4.2/rancheros-openstack.img
* https://releases.rancher.com/os/v1.4.2/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.4.2/rancheros-cloudstack.img
* https://releases.rancher.com/os/v1.4.2/rancheros-aliyun.vhd
* https://releases.rancher.com/os/v1.4.2/rancheros.ipxe
* https://releases.rancher.com/os/v1.4.2/rancheros-gce.tar.gz
* https://releases.rancher.com/os/v1.4.2/rootfs.tar.gz
* https://releases.rancher.com/os/v1.4.2/vmlinuz
* https://releases.rancher.com/os/v1.4.2/rancheros-vmware.iso

#### ARM Links

* https://releases.rancher.com/os/latest/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/latest/rancheros-raspberry-pi64.zip
* https://releases.rancher.com/os/v1.4.2/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v1.4.2/rancheros-raspberry-pi64.zip

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI
-------|------|------
ap-south-1 | HVM | [ami-0094b103f47719886](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-0094b103f47719886)
eu-west-3 | HVM | [ami-04902ea91b4485f74](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-04902ea91b4485f74)
eu-west-2 | HVM | [ami-0502a4b74a6e12b55](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-0502a4b74a6e12b55)
eu-west-1 | HVM | [ami-0c19809f0be281385](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-0c19809f0be281385)
ap-northeast-2 | HVM | [ami-01983b78f049ec6b4](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-01983b78f049ec6b4)
ap-northeast-1 | HVM | [ami-0bd72556695124d87](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-0bd72556695124d87)
sa-east-1 | HVM | [ami-04fa8430d60d238ce](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-04fa8430d60d238ce)
ca-central-1 | HVM | [ami-02f6436102b3c9c72](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-02f6436102b3c9c72)
ap-southeast-1 | HVM | [ami-0ce419a910d205864](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-0ce419a910d205864)
ap-southeast-2 | HVM | [ami-0dbd063b23c6d1c5d](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-0dbd063b23c6d1c5d)
eu-central-1 | HVM | [ami-0114900c022b09346](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-0114900c022b09346)
us-east-1 | HVM | [ami-08bb050b78c315da3](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-08bb050b78c315da3)
us-east-2 | HVM | [ami-02529740975197e75](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-02529740975197e75)
us-west-1 | HVM | [ami-094eaad86e2d89c00](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-094eaad86e2d89c00)
us-west-2 | HVM | [ami-03c3efd1e21bb4c6d](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-03c3efd1e21bb4c6d)
cn-north-1 | HVM | [ami-05003e029242a1f2a](https://cn-north-1.console.amazonaws.cn/ec2/home?region=cn-north-1#launchInstanceWizard:ami=ami-05003e029242a1f2a)
cn-northwest-1 | HVM | [ami-06b0f560c196abfe7](https://cn-northwest-1.console.amazonaws.cn/ec2/home?region=cn-northwest-1#launchInstanceWizard:ami=ami-06b0f560c196abfe7)

Additionally, images are available with support for Amazon EC2 Container Service (ECS) [here](https://rancher.com/docs/os/v1.x/en/installation/amazon-ecs/#amazon-ecs-enabled-amis).

### Google Compute Engine

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS.

[Download Latest Image](https://releases.rancher.com/os/latest/rancheros-gce.tar.gz)

[Download Stable Image](https://releases.rancher.com/os/v1.4.2/rancheros-gce.tar.gz)

Please follow the directions at our [docs to launch in GCE](https://rancher.com/docs/os/v1.x/en/installation/running-rancheros/cloud/gce/).

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](https://rancher.com/docs/os/v1.x/en/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

## Support, Discussion, and Community
If you need any help with RancherOS or Rancher, please join us at either our [Rancher forums](http://forums.rancher.com) or [#rancher IRC channel](http://webchat.freenode.net/?channels=rancher) where most of our team hangs out at.

For security issues, please email security@rancher.com instead of posting a public issue in GitHub.  You may (but are not required to) use the GPG key located on [Keybase](https://keybase.io/rancher).


Please submit any **RancherOS** bugs, issues, and feature requests to [rancher/os](//github.com/rancher/os/issues).

Please submit any **Rancher** bugs, issues, and feature requests to [rancher/rancher](//github.com/rancher/rancher/issues).

## License

Copyright (c) 2014-2018 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
