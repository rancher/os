---
title: Getting Started on Vagrant
layout: default

---

### Using Vagrant
The RancherOS ISO file can be loaded to KVM, Vmware, or VirtualBox and installed as a normal Linux virtual machine. We can use Vagrant to build a RancherOS virtual machine on any of these virtualization providers.

Vagrant can be downloaded and installed from [here](http://www.vagrantup.com/downloads.html), after installing Vagrant you should clone the RancherOS Vagrant [repository](https://github.com/rancherio/os-vagrant):

```sh
$ git clone https://github.com/rancherio/os-vagrant.git
$ cd os-vagrant
```

The Vagrantfile, which will be used to build a virtual machine on top of virtualbox will be like the following:


```sh
Vagrant.configure(2) do |config|
   config.vm.box       = "rancheros"
   config.vm.box_url   = "http://cdn.rancher.io/vagrant/x86_64/prod/rancheros_virtualbox.box"
   config.ssh.username = "rancher"

   config.vm.provider "virtualbox" do |vb|
      vb.check_guest_additions = false
      vb.functional_vboxsf     = false
      vb.memory = "1024"
      vb.gui = true
   end

   config.vm.synced_folder ".", "/vagrant", disabled: true
end
```
Vagrantfile will use the second version of the configuration, and it will specify the vagrant box url, and will deploy the ssh keys to rancher user. After that it will specify virtualbox as the provider, note that RancherOS needs at least **1GB** RAM.

And now run **vagrant up** which will import the vagrant box and create the virtual machine with RancherOS installed, you should see:

```
…
…
==> default: Machine booted and ready!
```