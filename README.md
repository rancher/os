# BurmillaOS

BurmillaOS is successor of [RancherOS](//github.com/rancher/os) which reached end of life.

![GitHub release](https://img.shields.io/github/v/release/burmilla/os.svg)
[![Docker Pulls](https://img.shields.io/docker/pulls/burmilla/os.svg)](https://store.docker.com/community/images/burmilla/os)
[![Go Report Card](https://goreportcard.com/badge/github.com/burmilla/os)](https://goreportcard.com/badge/github.com/burmilla/os)

The smallest, easiest way to run Docker in production at scale.  Everything in BurmillaOS is a container managed by Docker.  This includes system services such as udev and rsyslog.  BurmillaOS includes only the bare minimum amount of software needed to run Docker.  This keeps the binary download of BurmillaOS very small.  Everything else can be pulled in dynamically through Docker.

## How this works

Everything in BurmillaOS is a Docker container.  We accomplish this by launching two instances of
Docker.  One is what we call the system Docker which runs as the first process.  System Docker then launches
a container that runs the user Docker.  The user Docker is then the instance that gets primarily
used to create containers.  We created this separation because it seemed logical and also
it would really be bad if somebody did `docker rm -f $(docker ps -qa)` and deleted the entire OS.

![How it works](./howitworks.png "How it works")

## Documentation for BurmillaOS

Please refer to our [BurmillaOS Documentation](https://burmilla.github.io) website to read all about BurmillaOS. It has detailed information on how BurmillaOS works, getting-started and other details.

Please submit any **BurmillaOS** bugs, issues, and feature requests to [burmilla/os](//github.com/burmilla/os/issues).

## License

Copyright (c) 2020 Project Burmilla

Copyright (c) 2014-2020 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
