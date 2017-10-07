---
title: Changing Docker Versions in RancherOS

redirect_from:
  - os/v1.1/en/configuration/custom-docker/

---

## Changing Docker Versions

The version of User Docker used in RancherOS can be configured using a [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config) file or by using the `ros engine` command.

> **Note:** There are known issues in Docker when switching between versions. For production systems, we recommend setting the Docker engine only once [using a cloud-config](#setting-the-docker-engine-using-cloud-config).

### Available Docker engines

The `ros engine list` command can be used to show which Docker engines are available to switch to. This command will also provide details of which Docker engine is currently being used.

```
$ sudo ros engine list
disabled docker-1.10.3
disabled docker-1.11.2
current  docker-1.12.1
```

### Setting the Docker engine using cloud-config

RancherOS supports defining which Docker engine to use through the cloud-config file. To change the Docker version from the default packaged version, you can use the following cloud-config setting and select one of the available engines. In the following example, we'll use the cloud-config file to set RancherOS to use Docker 1.10.3 for User Docker.

```yaml
#cloud-config
rancher:
  docker:
    engine: docker-1.10.3
```

### Changing Docker engines after RancherOS has started

If you've already started RancherOS and want to switch Docker engines, you can change the Docker engine by using the `ros engine switch` command. In our example, we'll switch to Docker 1.11.2.

```
$ sudo ros engine switch docker-1.11.2
INFO[0000] Project [os]: Starting project
INFO[0000] [0/19] [docker]: Starting
Pulling docker (rancher/os-docker:1.11.2)...
1.11.2: Pulling from rancher/os-docker
2a6bbb293656: Pull complete
Digest: sha256:ec57fb24f6d4856d737e14c81a20f303afbeef11fc896d31b4e498829f5d18b2
Status: Downloaded newer image for rancher/os-docker:1.11.2
INFO[0007] Recreating docker
INFO[0007] [1/19] [docker]: Started
INFO[0007] Project [os]: Project started
$ docker version
Client:
 Version:      1.11.2
 API version:  1.23
 Go version:   go1.5.4
 Git commit:   b9f10c9
 Built:        Wed Jun  1 21:20:08 2016
 OS/Arch:      linux/amd64

Server:
 Version:      1.11.2
 API version:  1.23
 Go version:   go1.5.4
 Git commit:   b9f10c9
 Built:        Wed Jun  1 21:20:08 2016
 OS/Arch:      linux/amd64

```

### Enabling Docker engines

If you don't want to automatically switch Docker engines, you can also set which version of Docker to use after the next reboot by enabling a Docker engine.

```
$ sudo ros engine enable docker-1.10.3
```

## Using a Custom Version of Docker

If you're using a version of Docker that isn't available by default or a custom build of Docker then you can create a custom Docker image and service file to distribute it.

Docker engine images are built by adding the binaries to a folder named `engine` and then adding this folder to a `FROM scratch` image. For example, the following Dockerfile will build a Docker engine image.

```
FROM scratch
COPY engine /engine
```

Once the image is built a [system service]({{page.osbaseurl}}/system-services/adding-system-services/) configuration file must be created. An [example file](https://github.com/rancher/os-services/blob/master/d/docker-1.12.3.yml) can be found in the rancher/os-services repo. Change the `image` field to point to the Docker engine image you've built.

All of the previously mentioned methods of switching Docker engines are now available. For example, if your service file is located at `https://myservicefile` then the following cloud-config file could be used to use your custom Docker engine.

```yaml
#cloud-config
rancher:
  docker:
    engine: https://myservicefile
```
