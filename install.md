---
title: Installation
layout: default

---
#Installing RancherOS


## Installing to Disk

To install RancherOS on a new disk you can now use the `rancheros-install` command. 

For non-ec2 installs, before getting started create a cloud-init file that will provide your initial ssh keys. At a minimum something like:

```
#cloud-config
ssh_authorized_keys:
- ssh-rsa AAA... user@rancher
```

See section below for current supported cloud-init functionality.

The command arguments are as follows:

```
Usage:
rancheros-install [options]
Options:
-c cloud-config file
needed for SSH keys.
-d device
-f [ DANGEROUS! Data loss can happen ] partition/format without prompting
-t install-type:
generic
amazon-ebs
-v os-installer version.
-h print this
```

This command orchestrates installation from the rancher/os container. 

####Examples:
Virtualbox installation:

`sudo rancheros-install -d /dev/sda -c ./cloud_data.yml -v v0.1.1 -t generic`

## Using Vagrant
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
## Running RancherOS on AWS
RancherOS is available as an Amazon Web Services AMI, and can be easily run on EC2.  Let’s walk through how to import and create a RancherOS on EC2 machine:

1. First login to your AWS console, and on the EC2 dashboard, click on **“Launch a new instance”**:

![RancherOS on AWS 1]({{site.baseurl}}/img/Rancher_aws1.png)

2. Choose Community AMIs and search for RancherOS:

![RancherOS on AWS 2]({{site.baseurl}}/img/Rancher_aws2.png)

3. After configuring the network, size of the instance, and the security groups, review the information of the instance and choose a ssh key pair to be used with the EC2instance:

![RancherOS on AWS 3]({{site.baseurl}}/img/Rancher_aws3.png)

4. Download your new key pair, and then launch the instance, you should see the instance up and running:

![RancherOS on AWS 4]({{site.baseurl}}/img/Rancher_aws4.png)


If you prefer to use the AWS CLI the command below will launch a new instance using the RancherOS AMI: 

```sh
$ aws ec2 run-instances --image-id ami-eeaefc86 --count 1 \
--instance-type t1.micro --key-name MyKey --security-groups new-sg
```
where ami-eeaefc86 is the AMI of RancherOS in us-east-1 region. You can find the codes for AMIs in other regions on the RancherOS GitHub site. Now you can login to the RancherOS system:

```sh
$ ssh -i MyKey.pem rancher@<ip-of-ec2-instance>
[rancher@rancher ~]$
```




