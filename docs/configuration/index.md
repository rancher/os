---
title: Configuring RancherOS
layout: default

---

## Configuring RancherOS
---
The configuration of RancherOS is the compilation of different sources. It starts with a default configuration that is shipped with RancherOS, adds in anything found through cloud init process and finally includes any changes that have been made by the user. The cloud-init process can be found in more detail [here]({{site.baseurl}}/docs/cloud-config). 

Any user changes made to RancherOS are made by interacting with the `rancher.yml` file. If any values are changed from the default configuration, the new value is added to the `rancher.yml` file. More details on how to interact with the file can be found [here]({{site.baseurl}}/docs/rancher-yml).

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