---
title: Getting Started on Vagrant
layout: default

---

## Running RancherOS on Vagrant
---

We have created a [RancherOS Vagrant project](https://github.com/rancherio/os-vagrant) that allows you to quickly test out RancherOS.

Vagrant can be downloaded and installed from [here](http://www.vagrantup.com/downloads.html).

After installing Vagrant, you should clone the [RancherOS Vagrant repository](https://github.com/rancherio/os-vagrant). From the command line, go to the directory that you want to clone the repo into. Clone the repo and go into the newly cloned directory.

```bash
$ git clone https://github.com/rancherio/os-vagrant.git
$ cd os-vagrant
```

Within this directory, the Vagrantfile is the file that will be used to build a virtual machine on top of virtualbox. Vagrantfile will use the second version of the configuration, and it will specify the vagrant box url, and will deploy the ssh keys to rancher user. After that it will specify virtualbox as the provider, note that RancherOS needs at least **1GB** RAM.

This is what the file looks like:

```bash
# vi: set ft=ruby :

require_relative 'vagrant_rancheros_guest_plugin.rb'

# To enable rsync folder share change to false
$rsync_folder_disabled = true
$number_of_nodes = 1
$vm_mem = "1024"
$vb_gui = false


# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure(2) do |config|
    config.vm.box   = "rancherio/rancheros"
    config.vm.box_version = ">=0.2.0"

    (1..$number_of_nodes).each do |i|
        hostname = "rancher-%02d" % i

        config.vm.define hostname do |node|
            node.vm.provider "virtualbox" do |vb|
                vb.memory = $vm_mem
                vb.gui = $vb_gui
            end

        ip = "172.19.8.#{i+100}"
        node.vm.network "private_network", ip: ip

        # Disabling compression because OS X has an ancient version of rsync installed.
        # Add -z or remove rsync__args below if you have a newer version of rsync on your machine.
        node.vm.synced_folder ".", "/opt/rancher", type: "rsync",
            rsync__exclude: ".git/", rsync__args: ["--verbose", "--archive", "--delete", "--copy-links"],
            disabled: $rsync_folder_disabled

        end
    end
end
```

Run `vagrant up`. This will import the vagrant box and create the virtual machine with RancherOS installed. 

```bash
$ vagrant up
Bringing machine 'rancher-01' up with 'virtualbox' provider...
…
…
==> rancher-01: Machine booted and ready!
==> rancher-01: Configuring and enabling network interfaces...
```

## Logging into RancherOS
---

Now, let's log in to the system. We use `vagrant ssh` to authenticate with the private Vagrant key and login to the system:


```bash
$ vagrant ssh
```

After you're logged into the system, go back to the [Quick Start Guide]({{site.baseurl}}/docs/quick-start-guide/) to see some examples of what we can do.  

## Shutting Down the VM
---
If you are in the RancherOS command line, type `exit`.

```bash
$ exit 
logout
Connection to 127.0.0.1 closed. 
```

If you want to shut down your VM, run `vagrant halt` command from the os-vagrant directory. Or if you want to destroy the VM, run `vagrant destroy`. 

To get the VM back up, run `vagrant up` in the os-vagrant directory and just log back in `vagrant ssh`.

