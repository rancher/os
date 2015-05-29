---
title: Running RancherOS on VirtualBox using docker machine
layout: default

---

## Running RancherOS using Docker Machine on VirtualBox
---

### Requirements

1. VirtualBox in your path. Download and install from [VirtualBox Downloads page](https://www.virtualbox.org/wiki/Downloads)
2. Docker Machine. The machine version should be atleast [v0.3.0-rc1](https://github.com/docker/machine/releases/tag/v0.3.0-rc1). Download it from [the docker machine releases page](https://github.com/docker/machine/releases)

### Running RancherOS

1. Fetch the latest RancherOS ISO URL for machine from [the RancherOS releases page](https://github.com/rancherio/os/releases)

    At the time of writing this doc, the latest release is [v0.3.1](https://github.com/rancherio/os/releases/tag/v0.3.1)

    You might notice that there are two .iso files, rancheros.iso and machine-rancheros.iso.

    machine-rancheros.iso is the file you want. This ISO has been built with special configuration for setup with docker-machine.

    Copy the link to machine-rancheros.iso.

2. Go to your console, and use this command


        docker-machine create -d virtualbox --virtualbox-boot2docker-url $RANCHEROS-ISO-URL $MACHINE-NAME


    Note that, For the v0.3.1 release of RancherOS, the value for 
    
    ```RANCHEROS-ISO-URL``` is ```https://github.com/rancherio/os/releases/tag/v0.3.1```
    ```MACHINE-NAME``` is the name that you would like to call your machine. 

That's it. You have a RancherOS host running on virtualbox now. You can verify that you have a VirtualBox VM running on yor host using this command

``` VBoxManage list runningvms | grep $MACHINE-NAME```

It should print out the newly crated machine. If not, something went wrong with the provisioning step.

### Logging into RancherOS

Logging into RancherOS follows the standard docker-machine way. Use this command to login into your newly provisioned RancherOS VM.

```docker-machine ssh $MACHINE-NAME```

This will log you into the RancherOS VM. You'll then be able to explore the OS, run commands, spin up containers etc.

Once you've finished exploring, exit by pressing `Ctrl+D`

### Spinning up containers on RancherOS

You can point the docker client on your host to the docker daemon running inside of the VM. That way, you can run your docker commands like you had installed docker on your host. 

To point your docker client to the daemon inside the VM, use this command

```eval $(docker-machine env $MACHINE-NAME)```

You can run any docker commmand like you would normally, and it will execute your command in the RancherOS VM. 

### Running applications remotely using docker

```docker run -p 80:80 -p 443:443 -d nginx```

This will startup nginx on your VM. In order to access it, you need the ip address of the VM.

```docker-machine ip $MACHINE-NAME```

If you copy the IP address printed from the above command and paste it in your browser, you should see a Welcome Page for nginx!

