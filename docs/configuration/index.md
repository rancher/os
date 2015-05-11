---
title: Configuring RancherOS
layout: default

---

## Configuring RancherOS
---
The configuration of RancherOS is the compilation of different sources. It starts with a default configuration that is shipped with RancherOS, adds in anything found through cloud init process and finally includes any changes that have been made by the user. The cloud-init process can be found in more detail [here]({{site.baseurl}}/docs/cloud-config). Any changes made to RancherOS are made by interacting with the `rancher.yml` file. If any values are changed from the default configuration, the new value is added to the `rancher.yml` file. You can use `ros config` to edit and interact with the configuration. 

Here's a diagram of how the configuration of RancherOS is compiled.

![Configuration of RancherOS]({{site.baseurl}}/img/cloud-config.png)

We have various topics that cover how to configure specific areas of RancherOS.

[Networking]({{site.baseurl}}/configuration/networking/)<br>
[Users]({{site.baseurl}}/configuration/users/)<br>
[SSH Keys]({{site.baseurl}}/configuration/ssh-keys/)<br>
[Custom Console OS]({{site.baseurl}}/configuration/custom-console/)<br>
[Adding System Services]({{site.baseurl}}/configuration/system-services/)<br>
[Setting up Docker TLS]({{site.baseurl}}/configuration/setting-up-docker-tls/)<br>
[Loading Kernel Modules]({{site.baseurl}}/configuration/loading-kernel-modules/)<br>
[Installing Kernel Modules that require Kernel Headers]({{site.baseurl}}/configuration/kernel-modules-kernel-headers/)<br>
[DKMS]({{site.baseurl}}/configuration/dkms/)<br>
[Custom Kernels]({{site.baseurl}}/configuration/custom-kernels/)<br>
[Building custom RancherOS ISO]({{site.baseurl}}/configuration/custom-rancheros-iso/)<br>
[Pre-packing Docker Images]({{site.baseurl}}/configuration/prepacking-docker-images/)<br>

### ros
---

`ros` is the main command to interact with RancherOS configuration, here's the link to the [full ros config command docs]({{site.baseurl}}/docs/rancheros-tools/ros/config/). With these commands, you can get and set values in the configuration as well as import/export configurations.

_In v0.3.1+, we changed the command from `rancherctl` to `ros`._

## Cloud Config through Cloud-Init 

Before the first boot of the server, you might pass in a cloud config file to be processed through the cloud-init process. Please read the [cloud config section]({{site.baseurl}}/docs/cloud-config/) for more details on the supported cloud-init functionality in cloud-config.

## RancherOS Detailed Configuration 

Within RancherOS, there are various areas within the system that you might want to configure. Below, we'll outline the different keys that can be saved in the `rancher.yml` file. If you choose to change these settings in the cloud config with the cloud-init process, they must be within the `rancher` key. We cover these details in a separate [section]({{site.baseurl}}/docs/configuration/detailed-configuration/).
