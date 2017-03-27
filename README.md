# RancherOS

The smallest, easiest way to run Docker in production at scale.  Everything in RancherOS is a container managed by Docker.  This includes system services such as udev and rsyslog.  RancherOS includes only the bare minimum amount of software needed to run Docker.  This keeps the binary download of RancherOS very small.  Everything else can be pulled in dynamically through Docker.

## How this works

Everything in RancherOS is a Docker container.  We accomplish this by launching two instances of
Docker.  One is what we call the system Docker which runs as the first process.  System Docker then launches
a container that runs the user Docker.  The user Docker is then the instance that gets primarily
used to create containers.  We created this separation because it seemed logical and also
it would really be bad if somebody did `docker rm -f $(docker ps -qa)` and deleted the entire OS.

![How it works](docs/rancheros.png "How it works")

## Latest Release

**v0.9.1 - Docker 1.12.6 - Linux 4.9.18**

### ISO

https://releases.rancher.com/os/latest/rancheros.iso
https://releases.rancher.com/os/v0.9.1/rancheros.iso

### Additional Downloads

#### Latest Links

* https://releases.rancher.com/os/latest/initrd
* https://releases.rancher.com/os/latest/initrd-v0.9.1
* https://releases.rancher.com/os/latest/iso-checksums.txt
* https://releases.rancher.com/os/latest/rancheros-openstack.img
* https://releases.rancher.com/os/latest/rancheros.iso
* https://releases.rancher.com/os/latest/rancheros-v0.9.1.tar.gz
* https://releases.rancher.com/os/latest/rootfs.tar.gz
* https://releases.rancher.com/os/latest/vmlinuz
* https://releases.rancher.com/os/latest/vmlinuz-4.9.18-rancher

#### v0.9.1 Links

* https://releases.rancher.com/os/v0.9.1/initrd
* https://releases.rancher.com/os/v0.9.1/initrd-v0.9.1
* https://releases.rancher.com/os/v0.9.1/iso-checksums.txt
* https://releases.rancher.com/os/v0.9.1/rancheros-openstack.img
* https://releases.rancher.com/os/v0.9.1/rancheros.iso
* https://releases.rancher.com/os/v0.9.1/rancheros-v0.9.1.tar.gz
* https://releases.rancher.com/os/v0.9.1/rootfs.tar.gz
* https://releases.rancher.com/os/v0.9.1/vmlinuz
* https://releases.rancher.com/os/v0.9.1/vmlinuz-4.9.18-rancher

#### v0.9.0 Links

* https://releases.rancher.com/os/v0.9.0/rootfs_arm.tar.gz
* https://releases.rancher.com/os/v0.9.0/rootfs_arm64.tar.gz
* https://releases.rancher.com/os/v0.9.0/rancheros-raspberry-pi.zip

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI |
-------|------|------
ap-south-1 | HVM | [ami-fd1e6d92](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-fd1e6d92)
eu-west-2 | HVM | [ami-51776335](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-51776335)
eu-west-1 | HVM | [ami-481e232e](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-481e232e)
ap-northeast-2 | HVM | [ami-c32efdad](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-c32efdad)
ap-northeast-1 | HVM | [ami-33aaf154](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-33aaf154)
sa-east-1 | HVM | [ami-15ed8d79](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-15ed8d79)
ca-central-1 | HVM | [ami-e61fa282](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-e61fa282)
ap-southeast-1 | HVM | [ami-63b50900](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-63b50900)
ap-southeast-2 | HVM | [ami-86b7bbe5](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-86b7bbe5)
eu-central-1 | HVM | [ami-a71ecfc8](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-a71ecfc8)
us-east-1 | HVM | [ami-37b00f21](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-37b00f21)
us-east-2 | HVM | [ami-c61632a3](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-c61632a3)
us-west-1 | HVM | [ami-8998c3e9](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-8998c3e9)
us-west-2 | HVM | [ami-f6910496](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-f6910496)


### Google Compute Engine

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS.

[Download Image](https://github.com/rancher/os/releases/download/v0.9.1/rancheros-v0.9.1.tar.gz)

Please follow the directions at our [docs to launch in GCE](http://docs.rancher.com/os/running-rancheros/cloud/gce/).

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](http://docs.rancher.com/os/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

## Support, Discussion, and Community
If you need any help with RancherOS or Rancher, please join us at either our [Rancher forums](http://forums.rancher.com) or [#rancher IRC channel](http://webchat.freenode.net/?channels=rancher) where most of our team hangs out at.

For security issues, please email security@rancher.com instead of posting a public issue in GitHub.  You may (but are not required to) use the GPG key located on [Keybase](https://keybase.io/rancher).


Please submit any **RancherOS** bugs, issues, and feature requests to [rancher/os](//github.com/rancher/os/issues).

Please submit any **Rancher** bugs, issues, and feature requests to [rancher/rancher](//github.com/rancher/rancher/issues).

#License
Copyright (c) 2014-2017 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
