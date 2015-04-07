# rancherctl tls

`rancherctl tls` is used to generate both the client and server TLS certificates
for Docker.

## Enabling TLS

For TLS to be used with Docker you must first enable TLS, this can be done by doing

    sudo rancherctl config set user_docker.tls true
    sudo system-docker restart userdocker


## Sub commands

| Command  | Description                              |
|----------|------------------------------------------|
| `generate` | Generates client and server certificates |

## End to end example

### Enabled TLS for Docker

    sudo rancherctl config set user_docker.tls true

### Generate server certificate.

A server certificate must be generated for the hostname under which
you will access the server.  You can use an IP, "localhost", or "foo.example.com".

    sudo rancherctl tls generate -s --hostname localhost --hostname something.example.com
    sudo system-docker restart userdocker

### Generate client certificate

One or more client certificates must be generated so that you can access Docker

    sudo rancherctl tls generate
    sudo chown -R rancher ${HOME}/.docker

The above command will store the generated certificate in `${HOME}/.docker`.

### Test certificates

    export DOCKER_HOST=tcp://localhost:2376 DOCKER_TLS_VERIFY=1
    docker ps
