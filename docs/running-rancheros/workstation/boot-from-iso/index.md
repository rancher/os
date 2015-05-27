---
title: Booting from ISO
layout: default
---

## Booting from ISO
---
The RancherOS ISO file can be loaded to KVM, Vmware, or VirtualBox and installed as a normal Linux virtual machine. 

If you've chosen to use a different provisioner, you can download the rancheros.iso file from our [releases page](https://github.com/rancherio/os/releases/). You must boot with at least **1GB** of memory. If you boot with the ISO, the login is hard coded to **rancher/rancher**. Only the ISO has the password hard coded. If you run from a cloud or install to disk, SSH keys or a password of your choice is expected to be used.

## Persisting State
---
If you are running from the ISO, RancherOS will be running from memory. In order to persist to disk, you can format a file system with the label `RANCHER_STATE`. 

```bash
$ sudo mkfs.ext4 -L RANCHER_STATE /dev/xvda
```

`/dev/xvda` will be the disk that will hold the state.


### Example using VirtualBox


1. Download the RancherOS ISO.

2. Start up a VM from VirtualBox.
    
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
    ```

Next, read about how to [install to disk]({{site.baseurl}}/docs/running-rancheros/server/install-to-disk/) in order to have any changes to RancherOS to be saved.

<br>
