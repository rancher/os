---
title: Configuring TLS
layout: default

---

## RancherCTL TLS

`rancherctl tls` is used to generate both the client and server TLS certificates for Docker.

Remember, all `rancherctl` commands needs to be used with `sudo`. 

### Enabling TLS

For TLS to be used with Docker, you must first enable TLS, this can be done by doing these commands.

```bash
$ sudo rancherctl config set user_docker.tls true
$ sudo system-docker restart userdocker
```

### End to end example

#### Enable TLS for Docker

```bash
$ sudo rancherctl config set user_docker.tls true
$ sudo system-docker restart userdocker
```

#### Generate Server Certificate

A server certificate must be generated for the hostname under which you will access the server.  You can use an IP, "localhost", or "foo.example.com".

```bash
$ hostname
rancher
$ sudo rancherctl tls generate -s --hostname rancher
$ sudo system-docker restart userdocker
```

#### Generate Client Certificates

One or more client certificates must be generated so that you can access Docker. After generating the certificate, you can store the generated certificate in `${HOME}/.docker`.

```bash
$ sudo rancherctl tls generate -d /.docker
$ sudo chown -R rancher ${HOME}/.docker
```

#### Test certificates

```bash
$ export DOCKER_HOST=tcp://localhost:2376 DOCKER_TLS_VERIFY=1
$ docker ps
```
