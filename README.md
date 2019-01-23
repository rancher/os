# RancherOS

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

- **v1.5.0 - Docker 18.06.1-ce - Linux 4.14.85**

### ISO

- https://releases.rancher.com/os/v1.5.0/rancheros.iso
- https://releases.rancher.com/os/v1.5.0/hyperv/rancheros.iso
- https://releases.rancher.com/os/v1.5.0/4glte/rancheros.iso
- https://releases.rancher.com/os/v1.5.0/rancheros-vmware.iso [boot by docker-machine]
- https://releases.rancher.com/os/v1.5.0/vmware/rancheros.iso [boot from ISO]

### Additional Downloads

#### AMD64 Links

* https://releases.rancher.com/os/v1.5.0/initrd
* https://releases.rancher.com/os/v1.5.0/vmlinuz
* https://releases.rancher.com/os/v1.5.0/rancheros.ipxe
* https://releases.rancher.com/os/v1.5.0/rootfs.tar.gz

#### ARM64 Links

* https://releases.rancher.com/os/v1.5.0/arm64/initrd
* https://releases.rancher.com/os/v1.5.0/arm64/vmlinuz
* https://releases.rancher.com/os/v1.5.0/arm64/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v1.5.0/arm64/rancheros-raspberry-pi64.zip

#### Cloud Links

* https://releases.rancher.com/os/v1.5.0/rancheros-openstack.img
* https://releases.rancher.com/os/v1.5.0/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.5.0/rancheros-cloudstack.img
* https://releases.rancher.com/os/v1.5.0/rancheros-aliyun.vhd
* https://releases.rancher.com/os/v1.5.0/rancheros-gce.tar.gz

#### VMware Links

* https://releases.rancher.com/os/v1.5.0/vmware/initrd
* https://releases.rancher.com/os/v1.5.0/vmware/rancheros.vmdk
* https://releases.rancher.com/os/v1.5.0/vmware/rootfs.tar.gz

#### Hyper-V Links

* https://releases.rancher.com/os/v1.5.0/hyperv/initrd
* https://releases.rancher.com/os/v1.5.0/hyperv/rootfs.tar.gz

#### 4G-LTE Links

* https://releases.rancher.com/os/v1.5.0/4glte/initrd
* https://releases.rancher.com/os/v1.5.0/4glte/rootfs.tar.gz

**Note**:
1. you can use `http` instead of `https` in the above URLs, e.g. for iPXE.
2. you can use `latest` instead of `v1.5.0` in the above URLs if you want to get the latest version.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI
-------|------|------
ap-south-1 | HVM | [ami-0bcc62f1ebf353a70](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-0bcc62f1ebf353a70)
eu-west-3 | HVM | [ami-068d367d19530a393](https://eu-west-3.console.aws.amazon.com/ec2/home?region=eu-west-3#launchInstanceWizard:ami=ami-068d367d19530a393)
eu-west-2 | HVM | [ami-074223bf7b0ca9f7a](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-074223bf7b0ca9f7a)
eu-west-1 | HVM | [ami-00cd01b4ce6796af1](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-00cd01b4ce6796af1)
ap-northeast-2 | HVM | [ami-04c7c603a5586aafe](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-04c7c603a5586aafe)
ap-northeast-1 | HVM | [ami-09a1651831e76846a](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-09a1651831e76846a)
sa-east-1 | HVM | [ami-02703c46a20f47722](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-02703c46a20f47722)
ca-central-1 | HVM | [ami-06a0d5077db5cb530](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-06a0d5077db5cb530)
ap-southeast-1 | HVM | [ami-0443782221e818bd8](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-0443782221e818bd8)
ap-southeast-2 | HVM | [ami-031c628be26b5921f](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-031c628be26b5921f)
eu-central-1 | HVM | [ami-062a6985def70a2ca](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-062a6985def70a2ca)
us-east-1 | HVM | [ami-05342ca821afde9d7](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-05342ca821afde9d7)
us-east-2 | HVM | [ami-0be73aeb7d3076a36](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-0be73aeb7d3076a36)
us-west-1 | HVM | [ami-0052c7f3c5277f6b7](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-0052c7f3c5277f6b7)
us-west-2 | HVM | [ami-0b3a7af468ef99912](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-0b3a7af468ef99912)
cn-north-1 | HVM | [ami-032c116da4524fad2](https://cn-north-1.console.amazonaws.cn/ec2/home?region=cn-north-1#launchInstanceWizard:ami=ami-032c116da4524fad2)
cn-northwest-1 | HVM | [ami-08a6f3e3f9319127c](https://cn-northwest-1.console.amazonaws.cn/ec2/home?region=cn-northwest-1#launchInstanceWizard:ami=ami-08a6f3e3f9319127c)

Additionally, images are available with support for Amazon EC2 Container Service (ECS) [here](https://rancher.com/docs/os/v1.x/en/installation/amazon-ecs/#amazon-ecs-enabled-amis).

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
