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

**v0.2.1 - Docker 1.5.0 - Linux 3.19.2**

### ISO

https://github.com/rancherio/os/releases/download/v0.2.1/rancheros.iso

### Amazon

Region | Type | AMI |
-------|------|------
ap-northeast-1| PV | [ami-71cb3d71](https://console.aws.amazon.com/ec2/home?region=ap-northeast-1#launchAmi=ami-71cb3d71)
ap-southeast-1| PV | [ami-4a9eaf18](https://console.aws.amazon.com/ec2/home?region=ap-southeast-1#launchAmi=ami-4a9eaf18)
ap-southeast-2| PV | [ami-45ef9f7f](https://console.aws.amazon.com/ec2/home?region=ap-southeast-2#launchAmi=ami-45ef9f7f)
eu-west-1| PV | [ami-fd70ee8a](https://console.aws.amazon.com/ec2/home?region=eu-west-1#launchAmi=ami-fd70ee8a)
sa-east-1| PV | [ami-85f94298](https://console.aws.amazon.com/ec2/home?region=sa-east-1#launchAmi=ami-85f94298)
us-east-1| PV | [ami-5a321d32](https://console.aws.amazon.com/ec2/home?region=us-east-1#launchAmi=ami-5a321d32)
us-west-1| PV | [ami-bfa849fb](https://console.aws.amazon.com/ec2/home?region=us-west-1#launchAmi=ami-bfa849fb)
us-west-2| PV | [ami-a9bc9099](https://console.aws.amazon.com/ec2/home?region=us-west-2#launchAmi=ami-a9bc9099)

SSH keys are added to the **`rancher`** user, so you must log in using the **rancher** user.


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

