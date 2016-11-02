---
title: Configuring TLS in RancherOS
layout: os-default

---

## Setting up Docker TLS

`ros tls generate` is used to generate both the client and server TLS certificates for Docker.

Remember, all `ros` commands need to be used with `sudo` or as a `root` user.

### End to end example

#### Enable TLS for Docker and Generate Server Certificate

To have docker secured by TLS you need to set `rancher.docker.tls` to `true`, and generate a set of server and client keys and certificates:

```
$ sudo ros config set rancher.docker.tls true
$ sudo ros tls gen --server -H localhost -H <hostname1> -H <hostname2> ... -H <hostnameN>
$ sudo system-docker restart docker
```

Here, `<hostname*>`s are the hostnames that you will be able to use as your docker host names. A `<hostname*>` can be a wildcard pattern, e.g. "`*.*.*.*.*`". It is recommended to have `localhost` as one of the hostnames, so that you can test docker TLS connectivity locally.

When you've done that, all the necessary server certificate and key files have been saved to `/etc/docker/tls` directory, and the `docker` service has been started with `--tlsverify` option.

#### Generate Client Certificates

You also need client cert and key to access Docker via a TCP socket now:


```
$ sudo ros tls gen
  INFO[0000] Out directory (-d, --dir) not specified, using default: /home/rancher/.docker
```

All the docker client TLS files are in `~/.docker` dir now.

#### Test docker TLS connection

Now you can use your client cert to check if you can access Docker via TCP:

```
$ docker --tlsverify version
```

Because all the necessary files are in the `~/.docker` dir, you don't need to specify them using `--tlscacert` `--tlscert` and `--tlskey` options. You also don't need `-H` to access Docker on localhost.

Copy the files from `/home/rancher/.docker` to `$HOME/.docker` on your client machine if you need to access Docker on your RancherOS host from there.

On your client machine, set the Docker host and test out if Docker commands work.


```
$ export DOCKER_HOST=tcp://<hostname>:2376 DOCKER_TLS_VERIFY=1
$ docker ps
```
