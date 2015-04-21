---
title: Configuring TLS
layout: default

---

## RancherCTL TLS

`rancherctl tls` is used to generate both the client and server TLS certificates for Docker.

Remember, all `rancherctl` commands needs to be used with `sudo`. 

### End to end example

#### Enable TLS for Docker

```bash
$ sudo rancherctl config set user_docker.tls true
$ sudo system-docker restart userdocker
userdocker
```

#### Generate Server Certificate

A server certificate must be generated for the hostname under which you will access the server.  You can use an IP, "localhost", or "foo.example.com". If you want to see the certificate, use `rancherctl export config -p` to see all certificates.

```bash
$ hostname
rancher
$ sudo rancherctl tls generate -s --hostname rancher --hostname <IP_OF_SERVER>
$ sudo system-docker restart userdocker
userdocker
```

#### Generate Client Certificates

One or more client certificates must be generated so that you can access Docker. Let's store them in `/.docker` by using the `-d` option.

```bash
$ sudo rancherctl tls generate -d ~/.docker
# Change ownership to rancher user
$ sudo chown -R rancher .docker
```

After the certificates are created, you'll need to copy all 4 .pem files (ca-key.pem, ca.pem, cert.pem, key.pem) into your $HOME/.docker on your client machine. In this example, copy them to your local machine.

#### Test certificates

In your client, set the docker host and test out if Docker commands work. 

```bash
$ export DOCKER_HOST=tcp://<IP_OF_SERVER>:2376 DOCKER_TLS_VERIFY=1
$ docker ps
```
