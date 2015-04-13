---
title: Installing to Disk
layout: default
---

## Booting from ISO

The RancherOS ISO file can be loaded to KVM, Vmware, or VirtualBox and installed as a normal Linux virtual machine.

Download the rancherOS.iso file from our [releases page](https://github.com/rancherio/os/releases/).


### Using VirtualBox


1. Download the RancherOS ISO.

2. Start up a VM from one of the providers. In our example, we'll use VirtualBox.
    
    a. Open up VirtualBox. If you don't have VirtualBox, download it [here](https://www.virtualbox.org/wiki/Downloads).

     ![RancherOS on ISO 1]({{site.baseurl}}/img/Rancher_iso1.png)

    b. Provide a **name**, select the **type** to be _Linux_, and select the **version** to be _Other Linux (64-bit)_. Click **Continue**.
        
     ![RancherOS on ISO 2]({{site.baseurl}}/img/Rancher_iso2.png)

    c. Select at least **1GB** of RAM.

     ![RancherOS on ISO 3]({{site.baseurl}}/img/Rancher_iso3.png)

    d. Select **Create a virtual hard drive now** and click **Create**.

     ![RancherOS on ISO 4]({{site.baseurl}}/img/Rancher_iso4.png)

    e. Select the **VDI (VirtualBox Disk Image)** setting and click **Continue**.

     ![RancherOS on ISO 5]({{site.baseurl}}/img/Rancher_iso5.png)

    f. Select **Dynamically allocated** and click **Continue**.

     ![RancherOS on ISO 6]({{site.baseurl}}/img/Rancher_iso6.png)  

    g. Click **Create**.

     ![RancherOS on ISO 7]({{site.baseurl}}/img/Rancher_iso7.png)  
    
    Your new VM should be created, but in a _Powered Off_ state.

3. Start the VM from VirtualBox by clicking on the VM and clicking **Start** or right-click on the box and select **Start**. You will be immediately prompted to select an ISO. Find the RancherOS ISO that you have downloaded. Click **Start**.

    ![RancherOS on ISO 7]({{site.baseurl}}/img/Rancher_iso7.png)  

4. When RancherOS launches, you will be prompted for a rancher login and password. The login and password is 'rancher' (all lowercase).

    ```bash
    RancherOS rancher /dev/ttyl
    rancher login: rancher
    Password: 
    [rancher@rancher ~]$
    ```

5. Once you are logged into RancherOS, you can follow the rest of the [Getting Started Guide]({{site.baseurl}}/docs/getting-started/) to see what you can do. 

### Rebooting without Installing to Disk
When RancherOS is booted from an ISO, RancherOS is running from memory. Any changes in the console and containers will be lost upon rebooting.

Example:
In RancherOS, create a file with the word "test" in it.

```bash
[rancher@rancher ~]$ echo 'test' > MyFile
[rancher@rancher ~]$ ls
MyFile
[rancher@rancher ~]$ cat MyFile
test
[rancher@rancher ~]$
```

Reboot the VM.

```bash
[rancher@rancher ~]$ sudo reboot
```

When the VM is up and running again, you'll need to login again. Remember that the name and password are 'rancher' (all lowercase). Check to see if MyFile is still in the directory. 

```bash
[rancher@rancher ~]$ ls
[rancher@rancher ~]$
```

The file is no longer there. If you want this type of change to be saved upon rebooting, follow the steps in the next section to install to disk.


## Installing to Disk

To install RancherOS on a new disk, you can use the `rancheros-install` command. By installing to disk, this will allow any changes saved to the console and system containers to be saved whenever a reboot happens.

### Creating a Cloud Config file

Create a [cloud-init file](https://cloudinit.readthedocs.org/en/latest/index.html). If username and password are enabled, you don't need anything specific in your file. If username and password are not enabled, you should add your SSH public key. Make sure the file has a .yml extension.

Note: If you are trying to install Amazon EC2 to disk to create your own AMI, you will not need to create a cloud-init file.

Cloud-Init File Example:

```
#cloud-config
ssh_authorized_keys:
- ssh-rsa AAA... user@rancher
```

If you want more details on what Cloud Init functionality is supported, click [here]({{site.baseurl}}/docs/configuring/cloud-init/).  

For our VirtualBox example, let's create a cloud-init file named cloud_config.yml.

```bash
[rancher@rancher ~]$ echo '#cloud-config' > cloud_config.yml
```

### Using rancheros-install 

The `rancheros-install` command orchestrates installation from the rancher/os container. 

In continuing our VirtualBox installation, let's install RancherOS to disk, create a test file and reboot to see if the file is still there. 

```bash
[rancher@rancher ~]$ sudo rancheros-install -c cloud_config.yml -d /dev/sda -t generic -v v0.2.1
All data will be wiped from this device
Partition: true
DEVICE: /dev/sda
Are you sure you want to continue? [yN]
```
You will be prompted to see if you want to continue. Type **y**.

```
Are you sure you want to continue? [yN]yUnable to find image 'rancher/os:v0.2.1
Pulling repository rancher/os
...
...
...
Downloaded newer image for rancher/os:v0.2.1
+ DEVICE=/dev/sda
...
...
...
RancherOS has been installed. Please reboot...
[rancher@rancher ~]$ sudo reboot
```

Login to the VM instance using your SSH keys.

```bash
rancher login: rancher
Password:
[rancher@rancher ~]$ echo 'test' > MyFile
[rancher@rancher ~]$ ls
MyFile 
[rancher@rancher ~]$ cat MyFile
test
[rancher@rancher ~]$ sudo reboot

```

Log back into your RancherOS instance and your test file should still be there!

```bash
rancher login: rancher
Password:
[rancher@rancher ~]$ ls
MyFile
[rancher@rancher ~]$ 
```

Note: If you have chosen to use a different version from your iso during the install to disk, make sure you choose to boot off your primary disk instead of the ISO. Otherwise, you will be booted off your ISO version. In our VirtualBox example, that means you'd have to push F12 while booting.

#### Options

| Options | Description |
|--------|---------|
|-c | The Cloud-Config file needed for SSH keys |
| -d | Which Device to Install to |
|    -f | [ DANGEROUS! Data loss can happen ] Partition/Format without prompting |
|    -t | Decide the install-type: generic:    (Default) Creates 1 ext4 partition and installs RancherOS amazon-ebs: Installs RancherOS and sets up PV-GRUB
|    -v | Provide which os-installer version |
|    -h | Prints the help informations |

##### Cloud-Config

The `-c` option provides where to get the cloud config file. 

##### Device

The `-d` option determines where the disk will be installed to. 

In our VirtualBox example, you can see the list of disks available to install to. 

```bash
[rancher@rancher ~]$ sudo fdisk -l
Disk /dev/sda: 8589 MB, 8589934592 bytes
255 heads, 63 sectors/track, 1044 cylinders
Units = cylinders of 16065 *512 = 8225280 bytes

Disk /dev/sda doesn't contain a valid partition table
[rancher@rancher ~]$
```

##### Install Type
The `-t` option determines what type of installation is used. In most cases, we will be picking the _generic_ install type. The _amazon-ebs_ type is for creating your own AMI images. Since we are creating the [RancherOS AMI images]({{site.baseurl}}/docs/getting-started/amazon/), there is no need to create your own. 

##### Version

The `-v` option will indicate which version of RancherOS to install. To see the most recent versions of RancherOS, please vist the RancherOS GitHub page of [releases](https://github.com/rancherio/os/releases).

Alternatively, you can use the [rancherctl os]({{site.baseurl}}/docs/rancherctl/os/) commands to find the list of available versions.

```bash
[rancher@rancher ~]$ sudo rancherctl os list
rancher/os:v0.1.2 remote
rancher/os:v0.2.0-rc1 remote
rancher/os:v0.2.0-rc2 remote
rancher/os:v0.2.0-rc3 remote
rancher/os:v0.2.0 remote
rancher/os:v0.2.1 remote
rancher/os:v0.3.0-rc1 remote
[rancher@rancher ~]$ 
```




