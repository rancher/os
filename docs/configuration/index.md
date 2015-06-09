---
title: Configuring RancherOS
layout: default

---

## Configuring RancherOS
---
The configuration of RancherOS is derived from three sources.

1. RancherOS ships with a default configuration. Default configuration is hard coded into RancherOS binary. The default configuration cannot be changed, but it can be extended or overwritten by cloud-config and `rancher.yml`.

2. Cloud config extends and overwrites RancherOS default config. Cloud config itself is derived from several sources by the `cloud-init` program running as a system container inside RancherOS. The details of cloud config can be found in more detail [here]({{site.baseurl}}/docs/cloud-config). 

3. Finally, `rancher.yml` file extends and overwrites the result of cloud config. More details of the `rancher.yml` file can be found [here]({{site.baseurl}}/docs/rancher-yml).

The following diagram illustrates how RancherOS is configured from three sources: the default configuration, cloud config, and `rancher.yml` file.

![Configuration of RancherOS]({{site.baseurl}}/img/cloud-config.png)

You can see the RancherOs configuration in its entirety by typing `sudo ros config export --full`.

The following is a list of topics on RancherOS configuration:

[Networking]({{site.baseurl}}/docs/configuration/networking/)<br>
[Users]({{site.baseurl}}/docs/configuration/users/)<br>
[SSH Keys]({{site.baseurl}}/docs/configuration/ssh-keys/)<br>
[Custom Console OS]({{site.baseurl}}/docs/configuration/custom-console/)<br>
[Adding System Services]({{site.baseurl}}/docs/configuration/system-services/)<br>
[Setting up Docker TLS]({{site.baseurl}}/docs/configuration/setting-up-docker-tls/)<br>
[Loading Kernel Modules]({{site.baseurl}}/docs/configuration/loading-kernel-modules/)<br>
[Installing Kernel Modules that require Kernel Headers]({{site.baseurl}}/docs/configuration/kernel-modules-kernel-headers/)<br>
[DKMS]({{site.baseurl}}/docs/configuration/dkms/)<br>
[Custom Kernels]({{site.baseurl}}/docs/configuration/custom-kernels/)<br>
[Building custom RancherOS ISO]({{site.baseurl}}/docs/configuration/custom-rancheros-iso/)<br>
[Pre-packing Docker Images]({{site.baseurl}}/docs/configuration/prepacking-docker-images/)<br>
