---
title: Running RancherOS on Digital Ocean


---

## Running RancherOS on DigitalOcean
---

Running RancherOS on DigitalOcean is not yet supported, but there is a `rancheros` image now available from the commandline tools, so you can run:

```
$ doctl.exe compute droplet create --image rancheros --region sfo1 --size 2gb --ssh-keys 0a:db:77:92:03:b5:b2:94:96:d0:92:6a:e1:da:cd:28 myrancherosvm
ID          Name       Public IPv4    Private IPv4    Public IPv6    Memory    VCPUs    Disk    Region    Image                                    Status    Tags
47145723    myrancherosvm                                            2048      2        40      sfo1      RacherOS v1.0.1-rc [UNSUPPORTED/BETA]    new

$ doctl.exe compute droplet list
47145723    myrancherosvm                    107.170.203.111    10.134.26.83     2604:A880:0001:0020:0000:0000:2750:0001    2048      2        40      sfo1      RacherOS v1.0.1-rc [UNSUPPORTED/BETA]    active

ssh -i ~/.ssh/Sven.pem rancher@107.170.203.111
```

or use `docker-machine`:

```
$ docker-machine create -d digitalocean --digitalocean-access-token <your digital ocean token> --digitalocean-image rancheros --digitalocean-region sfo1 --digitalocean-size 2gb --digitalocean-ssh-user rancher sven-machine
Running pre-create checks...
Creating machine...
(sven-machine) Creating SSH key...
(sven-machine) Assuming Digital Ocean private SSH is located at ~/.ssh/id_rsa
(sven-machine) Creating Digital Ocean droplet...
(sven-machine) Waiting for IP address to be assigned to the Droplet...
Waiting for machine to be running, this may take a few minutes...
Detecting operating system of created instance...
Waiting for SSH to be available...
Detecting the provisioner...
Provisioning with rancheros...
Copying certs to the local machine directory...
Copying certs to the remote machine...
Setting Docker configuration on the remote daemon...
Checking connection to Docker...
Docker is up and running!
To see how to connect your Docker Client to the Docker Engine running on this virtual machine, run: C:\Users\svend\src\github.com\docker\machine\machine.exe env sven-machine
$ docker-machine ls
NAME            ACTIVE   DRIVER         STATE     URL                        SWARM   DOCKER        ERRORS
rancheros-100   -        virtualbox     Stopped                                      Unknown
sven-machine    -        digitalocean   Running   tcp://104.131.156.5:2376           v17.03.1-ce
$ docker-machine ssh sven-machine
Enter passphrase for key '/c/Users/svend/.ssh/id_rsa':
[rancher@sven-machine ~]$
[rancher@sven-machine ~]$
```
