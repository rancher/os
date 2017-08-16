---
title: Running RancherOS on Azure


---

## Running RancherOS on Azure
---

RancherOS is available as an image with Azure Resource Management. Please note that RancherOS is only offered in Azure Resource Management and not available in the Azure Service Management.

> **Note:** Currently, we only have v0.3.1 available as an image in Azure and it does not support passing in cloud config files. We are working on adding a new version that has cloud config enabled. Also, only certain regions are supported with RancherOS on Azure.

### Launching Rancheros through the Azure Portal

Using the new Azure Resource Management portal, click on **Marketplace**. Search for **RancherOS**. Click on **Create**.

Follow the steps to create a virtual machine.

In the _Basics_ step, provide a **name** for the VM, use _rancher_ as the **user name** and select the **SSH public key** option of authenticating. Add your ssh public key into the appropriate field. Select the **Resource group** that you want to add the VM to or create a new one. Select the **location** for your VM.

In the _Size_ step, select a virtual machine that has at least **1GB** of memory.

In the _Settings_ step, you can use all the default settings to get RancherOS running.

Review your VM and buy it so that you can **Create** your VM.

After the VM has been provisioned, click on the VM to find the public IP address. SSH into your VM using the _rancher_ username.

```
$ ssh rancher@<public_ip_of_vm> -p 22
```
