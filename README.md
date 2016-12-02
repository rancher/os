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

**v0.7.1 - Docker 1.12.3 - Linux 4.4.24**

### ISO

https://releases.rancher.com/os/latest/rancheros.iso  
https://releases.rancher.com/os/v0.7.1/rancheros.iso  

### Additional Downloads

#### Latest Links

##### v0.7.1
* https://releases.rancher.com/os/latest/initrd
* https://releases.rancher.com/os/latest/iso-checksums.txt
* https://releases.rancher.com/os/latest/rancheros-openstack.img
* https://releases.rancher.com/os/latest/rancheros.iso
* https://releases.rancher.com/os/latest/rancheros-v0.7.1.tar.gz
* https://releases.rancher.com/os/latest/rootfs.tar.gz
* https://releases.rancher.com/os/latest/vmlinuz

##### v0.7.0


* https://releases.rancher.com/os/latest/rancheros-raspberry-pi.zip
* https://releases.rancher.com/os/latest/rootfs_arm.tar.gz
* https://releases.rancher.com/os/latest/rootfs_arm64.tar.gz

#### v0.7.1 Links

* https://releases.rancher.com/os/v0.7.1/initrd
* https://releases.rancher.com/os/v0.7.1/iso-checksums.txt
* https://releases.rancher.com/os/v0.7.1/rancheros-openstack.img
* https://releases.rancher.com/os/v0.7.1/rancheros.iso
* https://releases.rancher.com/os/v0.7.1/rancheros-v0.7.1.tar.gz
* https://releases.rancher.com/os/v0.7.1/rootfs.tar.gz
* https://releases.rancher.com/os/v0.7.1/vmlinuz

#### v0.7.0 Links

* https://releases.rancher.com/os/v0.7.0/rancheros-raspberry-pi.zip
* https://releases.rancher.com/os/v0.7.0/rootfs_arm.tar.gz
* https://releases.rancher.com/os/v0.7.0/rootfs_arm64.tar.gz

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.  

### Amazon

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**HVM**

Region | Type | AMI |
-------|------|------
ap-northeast-1 | HVM |  [ami-be5bf2df](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-be5bf2df)
ap-northeast-2 | HVM |  [ami-247fab4a](https://console.aws.amazon.com/ec2/home?region=ap-northeast-2#launchInstanceWizard:ami=ami-247fab4a)
ap-south-1 | HVM |  [ami-dbf682b4](https://console.aws.amazon.com/ec2/home?region=ap-south-1#launchInstanceWizard:ami=ami-dbf682b4)
ap-southeast-1 | HVM |  [ami-c6d073a5](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-c6d073a5)
ap-southeast-2 | HVM |  [ami-51132c32](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-51132c32)
eu-central-1 | HVM |  [ami-2025df4f](https://console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-2025df4f)
eu-west-1 | HVM |  [ami-c62170b5](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-c62170b5)
sa-east-1 | HVM |  [ami-52b8273e](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-52b8273e)
us-east-1 | HVM |  [ami-dfdff3c8](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-dfdff3c8)
us-east-2 | HVM |  [ami-674c1602](https://console.aws.amazon.com/ec2/home?region=us-east-2#launchInstanceWizard:ami=ami-674c1602)
us-west-1 | HVM |  [ami-da2075ba](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-da2075ba)
us-west-2 | HVM |  [ami-ab3192cb](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-ab3192cb)

### Google Compute Engine

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS.

[Download Image](https://github.com/rancher/os/releases/download/v0.7.1/rancheros-v0.7.1.tar.gz)

Please follow the directions at our [docs to launch in GCE](http://docs.rancher.com/os/running-rancheros/cloud/gce/).

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](http://docs.rancher.com/os/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

## Support, Discussion, and Community
If you need any help with RancherOS or Rancher, please join us at either our [Rancher forums](http://forums.rancher.com) or [#rancher IRC channel](http://webchat.freenode.net/?channels=rancher) where most of our team hangs out at.

For issues relating to security, please email security@rancher.com instead of posting an open issue in Github.

Please submit any **RancherOS** bugs, issues, and feature requests to [rancher/os](//github.com/rancher/os/issues).

Please submit any **Rancher** bugs, issues, and feature requests to [rancher/rancher](//github.com/rancher/rancher/issues).

#License
Copyright (c) 2014-2016 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
