---
title: Booting from ISO
layout: default
---

## Booting from ISO
---
The RancherOS ISO file can be used to create a fresh RancherOS install on KVM, VMware, VirtualBox, or bare metal servers. You can download the `rancheros.iso` file from our [releases page](https://github.com/rancherio/os/releases/).

You must boot with at least **1GB** of memory. If you boot with the ISO, the login is hard coded to **rancher/rancher**. Only the ISO has the password hard coded. If you run from a cloud or install to disk, SSH keys or a password of your choice is expected to be used.

### Install to Disk
---
After you boot RancherOS from ISO, you can follow the instructions [here]({{site.baseurl}}/docs/running-rancheros/server/install-to-disk/) to install RancherOS to a hard disk.

### Persisting State
---
If you are running from the ISO, RancherOS will be running from memory. All downloaded Docker images, for example, will be stored in a ramdisk and will be lost after the server is rebooted. You can 
create a file system with the label `RANCHER_STATE` to instruct RancherOS to use that partition to store state. Suppose you have a disk partition on the server called /dev/sda, the following command formats that partition and labels it `RANCHER_STATE`

```bash
$ sudo mkfs.ext4 -L RANCHER_STATE /dev/sda
```

After you reboot the server RancherOS will use /dev/sda as the state partition.



<!----
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

---->
