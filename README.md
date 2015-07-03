# RancherOS

The smallest, easiest way to run Docker in production at scale.  Everything in RancherOS is a container managed by Docker.  This includes system services such as udev and rsyslog.  RancherOS includes only the bare minimum amount of software needed to run Docker.  This keeps the binary download of RancherOS to about 20MB.  Everything else can be pulled in dynamically through Docker.

## How this works

Everything in RancherOS is a Docker container.  We accomplish this by launching two instances of
Docker.  One is what we call the system Docker which runs as PID 1.  System Docker then launches
a container that runs the user Docker.  The user Docker is then the instance that gets primarily
used to create containers.  We created this separation because it seemed logical and also
it would really be bad if somebody did `docker rm -f $(docker ps -qa)` and deleted the entire OS.

![How it works](docs/rancheros.png "How it works")


## Latest Release

**v0.3.3 - Docker 1.7 - Linux 3.19.2**

### ISO

https://releases.rancher.com/os/latest/rancheros.iso  
https://releases.rancher.com/os/v0.3.3/rancheros.iso  

### Additional Downloads

https://releases.rancher.com/os/latest/machine-rancheros.iso  
https://releases.rancher.com/os/latest/iso-checksums.txt  
https://releases.rancher.com/os/latest/rancheros-033-gce-01.tar.gz  
https://releases.rancher.com/os/latest/vmlinuz  
https://releases.rancher.com/os/latest/initrd  

https://releases.rancher.com/os/v0.3.3/machine-rancheros.iso  
https://releases.rancher.com/os/v0.3.3/iso-checksums.txt  
https://releases.rancher.com/os/v0.3.3/rancheros-033-gce-01.tar.gz  
https://releases.rancher.com/os/v0.3.3/vmlinuz  
https://releases.rancher.com/os/v0.3.3/initrd  

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.  

### Amazon

We have 2 different [virtualization types of AMIs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/virtualization_types.html). SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**Paravirtual**

Region | Type | AMI |
-------|------|------
ap-northeast-1 | PV |  [ami-748a2774](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-748a2774)
ap-southeast-1 | PV |  [ami-3ae8e968](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-3ae8e968)
ap-southeast-2 | PV |  [ami-3531750f](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-3531750f)
eu-central-1 | PV |  [ami-d2b982cf](https://console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-d2b982cf)
eu-west-1 | PV |  [ami-fcb7f78b](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-fcb7f78b)
sa-east-1 | PV |  [ami-6361ec7e](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-6361ec7e)
us-east-1 | PV |  [ami-8f2eede4](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-8f2eede4)
us-west-1 | PV |  [ami-77a15133](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-77a15133)
us-west-2 | PV |  [ami-bfbfb98f](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-bfbfb98f)

**HVM**

HVM was introduced in v0.3.0 and only supports v0.3.0+.

Region | Type | AMI |
-------|------|------
ap-northeast-1 | HVM |  [ami-788a2778](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-788a2778)
ap-southeast-1 | HVM |  [ami-26e8e974](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-26e8e974)
ap-southeast-2 | HVM |  [ami-2b317511](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-2b317511)
eu-central-1 | HVM |  [ami-c8b982d5](https://console.aws.amazon.com/ec2/home?region=eu-central-1#launchInstanceWizard:ami=ami-c8b982d5)
eu-west-1 | HVM |  [ami-e0b7f797](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-e0b7f797)
sa-east-1 | HVM |  [ami-9f61ec82](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-9f61ec82)
us-east-1 | HVM |  [ami-772dee1c](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-772dee1c)
us-west-1 | HVM |  [ami-19a1515d](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-19a1515d)
us-west-2 | HVM |  [ami-c5bfb9f5](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-c5bfb9f5)

### Google Compute Engine (Experimental)

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS v0.3.0 or later.

[Download Image](https://github.com/rancherio/os/releases/download/v0.3.3/rancheros-033-gce-01.tar.gz)

Please follow the directions at our [docs to launch in GCE](http://os.docs.rancher.com/docs/running-rancheros/cloud/gce/). 

#### Known issues/ToDos
 * Add GCE daemon support. (Manages users)

## Documentation for RancherOS

Please refer to our [RancherOS Documentation](http://os.docs.rancher.com/) website to read all about RancherOS. It has detailed information on how RancherOS works, getting-started and other details.

## Support, Discussion, and Community
If you need any help with RancherOS or Rancher, please join us at either our [rancherio Google Groups](https://groups.google.com/forum/#!forum/rancherio) or [#rancher IRC channel](http://webchat.freenode.net/?channels=rancher) where most of our team hangs out at.

Please submit any **RancherOS** bugs, issues, and feature requests to [rancherio/os](//github.com/rancherio/os/issues).

Please submit any **Rancher** bugs, issues, and feature requests to [rancherio/rancher](//github.com/rancherio/rancher/issues).

#License
Copyright (c) 2014-2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

