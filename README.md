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

**v0.3.1 - Docker 1.6.2 - Linux 3.19.2**

### ISO

https://releases.rancher.com/os/latest/rancheros.iso  
https://releases.rancher.com/os/v0.3.1/rancheros.iso  

### Additional Downloads

https://releases.rancher.com/os/latest/machine-rancheros.iso  
https://releases.rancher.com/os/latest/iso-checksums.txt  
https://releases.rancher.com/os/latest/rancheros-031-gce-01.tar.gz  
https://releases.rancher.com/os/latest/vmlinuz  
https://releases.rancher.com/os/latest/initrd  

https://releases.rancher.com/os/v0.3.1/machine-rancheros.iso  
https://releases.rancher.com/os/v0.3.1/iso-checksums.txt  
https://releases.rancher.com/os/v0.3.1/rancheros-031-gce-01.tar.gz  
https://releases.rancher.com/os/v0.3.1/vmlinuz  
https://releases.rancher.com/os/v0.3.1/initrd  

**Note**: you can use `http` instead of `https` in the above URLs, e.g. for iPXE.  

### Amazon

We have 2 different [virtualization types of AMIs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/virtualization_types.html). SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**Paravirtual**

Region | Type | AMI |
-------|------|------
ap-northeast-1 | PV |  [ami-72fe2272](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-72fe2272)
ap-southeast-1 | PV |  [ami-e088b3b2](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-e088b3b2)
ap-southeast-2 | PV |  [ami-b989f183](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-b989f183)
eu-west-1 | PV |  [ami-993549ee](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-993549ee)
sa-east-1 | PV |  [ami-4fa02052](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-4fa02052)
us-east-1 | PV |  [ami-c78668ac](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-c78668ac)
us-west-1 | PV |  [ami-79a9433d](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-79a9433d)
us-west-2 | PV |  [ami-354c7505](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-354c7505)

**HVM**

HVM was introduced in v0.3.0 and only supports v0.3.0+.

Region | Type | AMI |
-------|------|------
ap-northeast-1 | HVM |  [ami-94fe2294](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-94fe2294)
ap-southeast-1 | HVM |  [ami-e888b3ba](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-e888b3ba)
ap-southeast-2 | HVM |  [ami-bf89f185](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-bf89f185)
eu-west-1 | HVM |  [ami-8b3549fc](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-8b3549fc)
sa-east-1 | HVM |  [ami-47a0205a](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-47a0205a)
us-east-1 | HVM |  [ami-818668ea](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-818668ea)
us-west-1 | HVM |  [ami-15a94351](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-15a94351)
us-west-2 | HVM |  [ami-114c7521](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-114c7521)


### Google Compute Engine (Experimental)

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS v0.3.0 or later.

[Download Image](https://github.com/rancherio/os/releases/download/v0.3.1/rancheros-031-gce-01.tar.gz)

#### Import
To import the image into your project follow the instructions below:

* [Upload an image into Compute Engine](https://cloud.google.com/compute/docs/tutorials/building-images#publishingimage)
* [Import RAW image](https://cloud.google.com/compute/docs/images#use_saved_image)


#### Usage
The image supports RancherOS cloud config functionality. Additionally, it merges the SSH keys from the project, instance and cloud-config and adds them to the **rancher** user.


#### Known issues/ToDos
 * Add GCE daemon support. (Manages users)


## Documentation for RancherOS

Please refer to our [website](http://rancherio.github.io/os/) to read all about RancherOS. It has detailed information on how it works, getting-started and other details.

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

