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

- **Latest: v1.4.0 - Docker 18.03.1-ce - Linux 4.14.32**
- **Stable: v1.4.0 - Docker 18.03.1-ce - Linux 4.14.32**

### ISO

- https://releases.rancher.com/os/latest/rancheros.iso
- https://releases.rancher.com/os/v1.4.0/rancheros.iso

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

#### v1.4.0 Links

* https://releases.rancher.com/os/v1.4.0/initrd
* https://releases.rancher.com/os/v1.4.0/iso-checksums.txt
* https://releases.rancher.com/os/v1.4.0/rancheros-openstack.img
* https://releases.rancher.com/os/v1.4.0/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.4.0/rancheros-cloudstack.img
* https://releases.rancher.com/os/v1.4.0/rancheros-aliyun.vhd
* https://releases.rancher.com/os/v1.4.0/rancheros.ipxe
* https://releases.rancher.com/os/v1.4.0/rancheros-gce.tar.gz
* https://releases.rancher.com/os/v1.4.0/rootfs.tar.gz
* https://releases.rancher.com/os/v1.4.0/vmlinuz
* https://releases.rancher.com/os/v1.4.0/rancheros-vmware.iso

#### ARM Links

* https://releases.rancher.com/os/latest/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/latest/rancheros-raspberry-pi64.zip
* https://releases.rancher.com/os/v1.4.0/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v1.4.0/rancheros-raspberry-pi64.zip

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI
-------|------|------
ap-south-1 | HVM | [ami-f4426c9b](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-f4426c9b)
eu-west-3 | HVM | [ami-6444f519](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-6444f519)
eu-west-2 | HVM | [ami-1e7f9379](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-1e7f9379)
eu-west-1 | HVM | [ami-447a7f3d](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-447a7f3d)
ap-northeast-2 | HVM | [ami-5492393a](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-5492393a)
ap-northeast-1 | HVM | [ami-96e218e9](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-96e218e9)
sa-east-1 | HVM | [ami-1a217876](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-1a217876)
ca-central-1 | HVM | [ami-eef6758a](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-eef6758a)
ap-southeast-1 | HVM | [ami-0716287b](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-0716287b)
ap-southeast-2 | HVM | [ami-4ae73528](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-4ae73528)
eu-central-1 | HVM | [ami-1686b3fd](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-1686b3fd)
us-east-1 | HVM | [ami-99c5ade6](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-99c5ade6)
us-east-2 | HVM | [ami-504b7435](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-504b7435)
us-west-1 | HVM | [ami-1e63797e](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-1e63797e)
us-west-2 | HVM | [ami-e59ae09d](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-e59ae09d)
cn-north-1 | HVM | [ami-0a5d8367](https://cn-north-1.console.amazonaws.cn/ec2/home?region=cn-north-1#launchInstanceWizard:ami=ami-0a5d8367)
cn-northwest-1 | HVM | [ami-40a1b522](https://cn-northwest-1.console.amazonaws.cn/ec2/home?region=cn-northwest-1#launchInstanceWizard:ami=ami-40a1b522)

Additionally, images are available with support for Amazon EC2 Container Service (ECS) [here](https://rancher.com/docs/os/v1.x/en/installation/amazon-ecs/#amazon-ecs-enabled-amis).

### Google Compute Engine

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS.

[Download Latest Image](https://releases.rancher.com/os/latest/rancheros-gce.tar.gz)

[Download Stable Image](https://releases.rancher.com/os/v1.4.0/rancheros-gce.tar.gz)

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
