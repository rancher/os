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

- **Latest: v1.3.0 - Docker 17.09.1-ce - Linux 4.9.80**
- **Stable: v1.2.0 - Docker 17.09.1-ce - Linux 4.9.78**

### ISO

- https://releases.rancher.com/os/latest/rancheros.iso
- https://releases.rancher.com/os/v1.2.0/rancheros.iso

### Additional Downloads

#### Latest Links

* https://releases.rancher.com/os/latest/initrd
* https://releases.rancher.com/os/latest/iso-checksums.txt
* https://releases.rancher.com/os/latest/rancheros-openstack.img
* https://releases.rancher.com/os/latest/rancheros-digitalocean.img
* https://releases.rancher.com/os/latest/rancheros-aliyun.vhd
* https://releases.rancher.com/os/latest/rancheros.ipxe
* https://releases.rancher.com/os/latest/rancheros-gce.tar.gz
* https://releases.rancher.com/os/latest/rootfs.tar.gz
* https://releases.rancher.com/os/latest/vmlinuz

#### v1.2.0 Links

* https://releases.rancher.com/os/v1.2.0/initrd
* https://releases.rancher.com/os/v1.2.0/iso-checksums.txt
* https://releases.rancher.com/os/v1.2.0/rancheros-openstack.img
* https://releases.rancher.com/os/v1.2.0/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.2.0/rancheros-aliyun.vhd
* https://releases.rancher.com/os/v1.2.0/rancheros.ipxe
* https://releases.rancher.com/os/v1.2.0/rancheros-gce.tar.gz
* https://releases.rancher.com/os/v1.2.0/rootfs.tar.gz
* https://releases.rancher.com/os/v1.2.0/vmlinuz

#### ARM Links

* https://releases.rancher.com/os/latest/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/latest/rancheros-raspberry-pi64.zip
* https://releases.rancher.com/os/v1.2.0/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v1.2.0/rancheros-raspberry-pi64.zip

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI(Stable) | AMI(Latest)
-------|------|------|------
ap-south-1 | HVM | [ami-12db887d](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-12db887d) | [ami-7e227811](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-7e227811)
eu-west-3 | HVM | [ami-d5a315a8](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-d5a315a8) | [ami-662d9b1b](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-662d9b1b)
eu-west-2 | HVM | [ami-80bd58e7](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-80bd58e7) | [ami-89bb5aee](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-89bb5aee)
eu-west-1 | HVM | [ami-69187010](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-69187010) | [ami-386b3841](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-386b3841)
ap-northeast-2 | HVM | [ami-57dd7f39](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-57dd7f39) | [ami-7da90613](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-7da90613)
ap-northeast-1 | HVM | [ami-a3c2b5c5](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-a3c2b5c5) | [ami-14070f68](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-14070f68)
sa-east-1 | HVM | [ami-6c2f6100](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-6c2f6100) | [ami-7345121f](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-7345121f)
ca-central-1 | HVM | [ami-b8a622dc](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-b8a622dc) | [ami-1ea3257a](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-1ea3257a)
ap-southeast-1 | HVM | [ami-0f5a1b73](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-0f5a1b73) | [ami-4b4d1437](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-4b4d1437)
ap-southeast-2 | HVM | [ami-edc73c8f](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-edc73c8f) | [ami-e4498586](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-e4498586)
eu-central-1 | HVM | [ami-28422647](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-28422647) | [ami-bc386557](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-bc386557)
us-east-1 | HVM | [ami-a7151cdd](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-a7151cdd) | [ami-38964e45](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-38964e45)
us-east-2 | HVM | [ami-a383b6c6](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-a383b6c6) | [ami-a0d7e6c5](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-a0d7e6c5)
us-west-1 | HVM | [ami-c4b3bca4](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-c4b3bca4) | [ami-a4a0b6c4](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-a4a0b6c4)
us-west-2 | HVM | [ami-6e1a9e16](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-6e1a9e16) | [ami-c60c94be](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-c60c94be)
cn-north-1 | HVM | [ami-18c11e75](https://cn-north-1.console.amazonaws.cn/ec2/home?region=cn-north-1#launchInstanceWizard:ami=ami-18c11e75) | [ami-07a37d6a](https://cn-north-1.console.amazonaws.cn/ec2/home?region=cn-north-1#launchInstanceWizard:ami=ami-07a37d6a)
cn-northwest-1 | HVM | [ami-bcedf9de](https://cn-northwest-1.console.amazonaws.cn/ec2/home?region=cn-northwest-1#launchInstanceWizard:ami=ami-bcedf9de) | [ami-8f8094ed](https://cn-northwest-1.console.amazonaws.cn/ec2/home?region=cn-northwest-1#launchInstanceWizard:ami=ami-8f8094ed)

Additionally, images are available with support for Amazon EC2 Container Service (ECS) [here](https://docs.rancher.com/os/amazon-ecs/#amazon-ecs-enabled-amis).

### Google Compute Engine

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS.

[Download Latest Image](https://releases.rancher.com/os/latest/rancheros-gce.tar.gz)

[Download Stable Image](https://releases.rancher.com/os/v1.2.0/rancheros-gce.tar.gz)

Please follow the directions at our [docs to launch in GCE](http://docs.rancher.com/os/running-rancheros/cloud/gce/).

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](http://docs.rancher.com/os/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

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
