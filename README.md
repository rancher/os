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

**v0.3.0 - Docker 1.6.0 - Linux 3.19.2**

### ISO

https://github.com/rancherio/os/releases/download/v0.3.0/rancheros.iso

### Amazon

We have 2 different [virtualization types of AMIs](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/virtualization_types.html). SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.

**Paravirtual**

Region | Type | AMI |
-------|------|------
ap-northeast-1 | PV |  [ami-506da950](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-506da950)
ap-southeast-1 | PV |  [ami-14043946](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-14043946)
ap-southeast-2 | PV |  [ami-37acd10d](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-37acd10d)
eu-west-1 | PV |  [ami-2d8fec5a](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-2d8fec5a)
sa-east-1 | PV |  [ami-fd7ffae0](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-fd7ffae0)
us-east-1 | PV |  [ami-ca3e39a2](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-ca3e39a2)
us-west-1 | PV |  [ami-55816c11](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-55816c11)
us-west-2 | PV |  [ami-29c1f419](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-29c1f419)


**HVM**

HVM was introduced in v0.3.0 and only supports v0.3.0+.

Region | Type | AMI |
-------|------|------
ap-northeast-1 | HVM |  [ami-be6da9be](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchInstanceWizard:ami=ami-be6da9be)
ap-southeast-1 | HVM |  [ami-06043954](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchInstanceWizard:ami=ami-06043954)
ap-southeast-2 | HVM |  [ami-23acd119](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchInstanceWizard:ami=ami-23acd119)
eu-west-1 | HVM |  [ami-018fec76](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchInstanceWizard:ami=ami-018fec76)
sa-east-1 | HVM |  [ami-f17ffaec](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchInstanceWizard:ami=ami-f17ffaec)
us-east-1 | HVM |  [ami-b03e39d8](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchInstanceWizard:ami=ami-b03e39d8)
us-west-1 | HVM |  [ami-6d816c29](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchInstanceWizard:ami=ami-6d816c29)
us-west-2 | HVM |  [ami-1fc1f42f](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchInstanceWizard:ami=ami-1fc1f42f)


### Google Compute Engine (Experimental)

We are providing a disk image that users can download and import for use in Google Compute Engine. The image can be obtained from the release artifacts for RancherOS v0.3.0.

[Download Image](https://github.com/rancherio/os/releases/download/v0.3.0/rancheros-030-gce-1.tar.gz)

#### Import
To import the image into your project follow the instructions below:

* [Upload an image into Compute Engine](https://cloud.google.com/compute/docs/tutorials/building-images#publishingimage)
* [Import RAW image](https://cloud.google.com/compute/docs/images#use_saved_image)


#### Usage
The image supports RancherOS cloud config functionality. Additionally, it merges the SSH keys from the project, instance and cloud-config and adds them to the **rancher** user.


#### Known issues/ToDos
 * Add GCE daemon support. (Manages users)


## Documentation for Rancher Labs

Please refer to our [website](http://rancherio.github.io/os/) to read all about RancherOS. It has detailed information on how it works, getting-started and other details.


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

