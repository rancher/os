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

**v1.1.1 - Docker 17.03.2-ce - Linux 4.9.67**

### ISO

- https://releases.rancher.com/os/latest/rancheros.iso
- https://releases.rancher.com/os/v1.1.1/rancheros.iso

### Additional Downloads

#### Latest Links

* https://releases.rancher.com/os/latest/initrd
* https://releases.rancher.com/os/latest/iso-checksums.txt
* https://releases.rancher.com/os/latest/rancheros-openstack.img
* https://releases.rancher.com/os/latest/rancheros-digitalocean.img
* https://releases.rancher.com/os/latest/rancheros.ipxe
* https://releases.rancher.com/os/latest/rancheros-gce.tar.gz
* https://releases.rancher.com/os/latest/rootfs.tar.gz
* https://releases.rancher.com/os/latest/vmlinuz

#### v1.1.1 Links

* https://releases.rancher.com/os/v1.1.1/initrd
* https://releases.rancher.com/os/v1.1.1/iso-checksums.txt
* https://releases.rancher.com/os/v1.1.1/rancheros-openstack.img
* https://releases.rancher.com/os/v1.1.1/rancheros-digitalocean.img
* https://releases.rancher.com/os/v1.1.1/rancheros.ipxe
* https://releases.rancher.com/os/v1.1.1/rancheros-gce.tar.gz
* https://releases.rancher.com/os/v1.1.1/rootfs.tar.gz
* https://releases.rancher.com/os/v1.1.1/vmlinuz

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI |
-------|------|------
ap-south-1 | HVM | [ami-97ade5f8](https://ap-south-1.console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-97ade5f8)
eu-west-2 | HVM | [ami-82a5bbe6](https://eu-west-2.console.aws.amazon.com/ec2/home?region=eu-west-2#launchInstanceWizard:ami=ami-82a5bbe6)
eu-west-1 | HVM | [ami-c5b134bc](https://eu-west-1.console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-c5b134bc)
ap-northeast-2 | HVM | [ami-6c54f202](https://ap-northeast-2.console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-6c54f202)
ap-northeast-1 | HVM | [ami-ece6648a](https://ap-northeast-1.console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-ece6648a)
sa-east-1 | HVM | [ami-0750166b](https://sa-east-1.console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-0750166b)
ca-central-1 | HVM | [ami-1cdd6778](https://ca-central-1.console.aws.amazon.com/ec2/home?region=ca-central-1#launchInstanceWizard:ami=ami-1cdd6778)
ap-southeast-1 | HVM | [ami-f4e58488](https://ap-southeast-1.console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-f4e58488)
ap-southeast-2 | HVM | [ami-1544b377](https://ap-southeast-2.console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-1544b377)
eu-central-1 | HVM | [ami-b632bad9](https://eu-central-1.console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-b632bad9)
us-east-1 | HVM | [ami-4584ec3f](https://us-east-1.console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-4584ec3f)
us-east-2 | HVM | [ami-4695bc23](https://us-east-2.console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-4695bc23)
us-west-1 | HVM | [ami-d24e4ab2](https://us-west-1.console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-d24e4ab2)
us-west-2 | HVM | [ami-99cb6de1](https://us-west-2.console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-99cb6de1)

Additionally, images are available with support for Amazon EC2 Container Service (ECS) [here](https://docs.rancher.com/os/amazon-ecs/#amazon-ecs-enabled-amis).

### Google Compute Engine

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS.

[Download Image](https://releases.rancher.com/os/v1.1.1/rancheros-gce.tar.gz)

Please follow the directions at our [docs to launch in GCE](http://docs.rancher.com/os/running-rancheros/cloud/gce/).

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](http://docs.rancher.com/os/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

## Support, Discussion, and Community
If you need any help with RancherOS or Rancher, please join us at either our [Rancher forums](http://forums.rancher.com) or [#rancher IRC channel](http://webchat.freenode.net/?channels=rancher) where most of our team hangs out at.

For security issues, please email security@rancher.com instead of posting a public issue in GitHub.  You may (but are not required to) use the GPG key located on [Keybase](https://keybase.io/rancher).


Please submit any **RancherOS** bugs, issues, and feature requests to [rancher/os](//github.com/rancher/os/issues).

Please submit any **Rancher** bugs, issues, and feature requests to [rancher/rancher](//github.com/rancher/rancher/issues).

## License

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
