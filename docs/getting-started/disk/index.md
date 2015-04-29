---
title: Installing to Disk
layout: default
---

## Booting from ISO
---
The RancherOS ISO file can be loaded to KVM, Vmware, or VirtualBox and installed as a normal Linux virtual machine. [Vagrant]({{site.baseurl}}/docs/getting-started/vagrant/) simplest way to try out RancherOS is using our [RancherOS Vagrant project](https://github.com/rancherio/os-vagrant).


If you've chosen to use a different provisioner, you can download the rancherOS.iso file from our [releases page](https://github.com/rancherio/os/releases/). You must boot with at least **1GB** of memory. If you boot with the ISO, the login is hard coded to **rancher/rancher**. Only the ISO has the password hard coded. If you run from a cloud or install to disk, SSH keys or a password of your choice is expected to be used.

### Example using VirtualBox


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

## Persisting State
---
If you are running from the ISO, RancherOS will be running from memory. In order to persist to disk, you can format a file system with the label `RANCHER_STATE`. 

Example:

```
docker run --privileged -it debian mkfs.ext4 -L RANCHER_STATE /dev/sda
```


## Installing to Disk
---
To install RancherOS on a new disk, you can use the `rancheros-install` command. By installing to disk, this will allow any changes saved to the console and system containers to be saved whenever a reboot happens.

### Adding Port Forwarding

Currently, RancherOS will dynamically reset the IP every time it reboots, so we will need to add a port forwarding rule so that we can SSH into our VM. Let's continue with our VirtualBox example.

We'll need to power off our VM in order to make our port forwarding rule. Select the VM, click on **Settings** and navigate to the **Network** tab. Expand the **Advanced** section and click on the **Port Forwarding** button.

 ![RancherOS to Disk 1]({{site.baseurl}}/img/Rancher_disk1.png)  

Click on the **+** icon to add a new port forwarding rule. Create the rule with the following parameters:

Name: ssh; Protocol: TCP; Host IP: 127.0.0.1; Host Port: 2223; Guest IP: <blank>; Guest Port: 22

Click on **OK**.

 ![RancherOS to Disk 2]({{site.baseurl}}/img/Rancher_disk2.png)

**Start** the VM and login to the VM. Username and password is 'rancher' (all lowercase).

### Creating a Cloud Config file

We will need to create a [cloud-init file](https://cloudinit.readthedocs.org/en/latest/index.html) that includes the public SSH key from your local computer. If you want more details on what Cloud Init functionality is supported, click [here]({{site.baseurl}}/docs/configuring/cloud-init/).  

Note: If you are trying to install Amazon EC2 to disk to create your own RancherOS AMI, you will not need to create a cloud-init file.

1. Check to see if you have SSH keys on your local computer. If not, generate a new one by following this [article](https://help.github.com/articles/generating-ssh-keys/). 

    ```bash
    $ ls -al ~/.ssh
    # Lists all the files in your .ssh directory, if they exist
    ```
    You are looking to see if you have one of the following:

    ```bash
    id_dsa.pub
    id_ecdsa.pub
    id_ed25519.pub
    id_rsa.pub
    ```

2. After you've ensured that you have a public SSH key, we'll proceed with copyubg the public SSH key from our computer to our VM. In our example, the public key that we're using is `id_rsa.pub`.

    ```bash
    [rancher@rancher ~]$ scp -r computer_username@computer_ip:~/.ssh/id_rsa.pub ./
    RSA key fingerprint is X:X:X:X:X.
    Are you sure you want to continue connecting (yes/no)? yes
    Warning: Permanently added 'computer_ip' (RSA) to the list of known hosts. 
    Password: 
    id_rsa.pub                              100%    422 0.4KB/s    00:00
    [rancher@rancher ~]$ ls
    id_rsa.pub
    [rancher@rancher ~]$ cat id_rsa.pub > cloud_config.yml
    [rancher@rancher ~]$ vi cloud_config.yml
    ```

3. Let's edit the cloud_config.yml so that it matches the syntax of the cloud-init example. Yaml files are very particular about their white spacing, so please note the spaces in our file!

    Cloud-Init File Example:

    ```yaml
    #cloud-config
    
    ssh_authorized_keys:
    - ssh-rsa AAA... user@host
    ```

Now that our cloud_config.yml contains our public SSH key, we can move on to installing RancherOS to disk!

### Using rancheros-install 

The `rancheros-install` command orchestrates the installation from the rancher/os container. We will install RancherOS to disk, reboot and then try to save a file. 

```bash
[rancher@rancher ~]$ sudo rancheros-install -c cloud_config.yml -d /dev/sda -v v0.2.1
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
```

In our VirtualBox setup, the VM will always boot from ISO, unless you change the boot order or remove the ISO. It's easiest to remove the ISO so that we don't have to worry about boot order every time we reboot.

Let's power off our VM to make the change regarding the ISO. Select the VM and click on **Settings**. Click on the **Storage** tab. Select the **rancheros.iso** file from the storage tree and click on the **delete** icon. 

It will confirm that you want to remove the CD/DVD device. Click on **Remove** and then click **OK**.  Start your VM. Since you have removed the ISO from the VM, it will select the current OS version running on the VM. 

![RancherOS to Disk 3]({{site.baseurl}}/img/Rancher_disk3.png)


Even though you will be prompted with the "rancher login" from the VirtualBox screen, you will not be able to use the previous rancher login/password. Instead, you will need to log in to the VM using your SSH keys.

Note: Once you've installed to disk, you will always need to log in as the **rancher** user. All SSH keys are passed to this user, so please make sure to ssh as the **rancher** user. 

Open up your terminal/command line. After we SSH in, let's create our file and reboot the VM to see if the file has been saved.

```bash
$ ssh rancher@127.0.0.1 -p 2223
[rancher@rancher ~]$ echo 'test' > MyFile
[rancher@rancher ~]$ ls
MyFile 
[rancher@rancher ~]$ cat MyFile
test
[rancher@rancher ~]$ sudo reboot

```

Wait a little bit before attempting to SSH back into your RancherOS instance and your test file should still be there. 

```bash
$ ssh rancher@127.0.0.1 -p 2223
[rancher@rancher ~]$ ls
MyFile
[rancher@rancher ~]$ 
```

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

By default, the install type will be generic.

##### Version

The `-v` option will indicate which version of RancherOS to install. To see the most recent versions of RancherOS, please vist the RancherOS GitHub page of [releases](https://github.com/rancherio/os/releases).

By default, the version installed will be version that RancherOS is currently running.

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

## Setting up a State Partition

Another way to save the state of Rancher OS is to create a state partition labeled as `RANCHER_STATE`. We package mkfs.ext4 in the console. 

```bash
$ sudo mkfs.ext4 -L RANCHER_STATE /dev/xvda
```

`/dev/xvda` will be the disk that will hold the state.

