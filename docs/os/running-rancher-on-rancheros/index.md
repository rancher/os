---
title: Running Rancher on RancherOS
layout: os-default

---

## Tips on using Rancher with RancherOS
---

RancherOS can be used to launch [Rancher]({{site.baseurl}}/rancher/) and be used as the OS to [add hosts]({{site.baseurl}}/rancher/rancher-ui/infrastructure/hosts/custom) to Rancher.

### Launching Agents using Cloud-Config

You can easily add hosts into Rancher by using [cloud-config]({{site.baseurl}}/os/configuration/#cloud-config) to launch the rancher/agent container. 

After Rancher is launched and [host registration]({{site.baseurl}}/rancher/configuration/settings/#host-registration) has been saved, you will be able to find the [custom command]({{site.baseurl}}/rancher/rancher-ui/infrastructure/hosts/custom) in the **Infrastructure** -> **Hosts** -> **Custom** page. 

```bash
$ sudo docker run --d --privileged -v /var/run/docker.sock:/var/run/docker.sock \
    rancher/agent:v0.8.2  http://<rancher-server-ip>:8080/v1/projects/1a5/scripts/<registrationToken>
```

<br>

> **Note:** The `rancher/agent` version is correlated to the Rancher server version. You will need to check the custom command to get the appropriate tag for the version to use.

_Cloud-Config Example_

Here's using the command above and converting it into a cloud-config file to launch the rancher/agent in docker when RancherOS boots up.

```yaml
#cloud-config
rancher:
  services:
    rancher-agent1:
      image: rancher/agent:v0.8.2
      command: http://<rancher-server-ip>:8080/v1/projects/1a5/scripts/<registrationToken>
      privileged: true
      volumes:
        - /var/run/docker.sock:/var/run/docker.sock
```
<br>

> **Note:** You can not name the service `rancher-agent` as this will not allow the rancher/agent container to be launched correctly. Please read more about why [you can't name your container as `rancher-agent`]({{site.baseurl}}/rancher/faqs/agents/#adding-in-name-rancher-agent). 

### Adding in Host Labels

With each host, you have the ability to add labels to help you organize your hosts. The labels are added as an environment variable when launching the rancher/agent container. The host label in the UI will be a key/value pair and the keys must be unique identifiers. If you added two keys with different values, we'll take the last inputted value to use as the key/value pair.

By adding labels to hosts, you can use these labels when [schedule services/load balancers/services]({{site.baseurl}}/rancher/rancher-ui/scheduling/) and create a whitelist or blacklist of hosts for your [services]({{site.baseurl}}/rancher/rancher-ui/applications/stacks/adding-services/) to run on. 

When adding a custom host, you can add the labels using the UI and it will automatically add the environment variable (`CATTLE_HOST_LABELS`) with the key/value pair into the command on the UI screen.

#### Native Docker Commands Example

```bash
# Adding one host label to the rancher/agent command
$  sudo docker run -e CATTLE_HOST_LABELS='foo=bar' -d --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock rancher/agent:v0.8.2 \
  http://<rancher-server-ip>:8080/v1/projects/1a5/scripts/<registrationToken>

# Adding more than one host label requires joining the additional host labels with an `&`
$  sudo docker run -e CATTLE_HOST_LABELS='foo=bar&hello=world' -d --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock rancher/agent:v0.8.2 \
  http://<rancher-server-ip>:8080/v1/projects/1a5/scripts/<registrationToken>
```

#### Cloud-Config Example

Adding one host label

```yaml
#cloud-config
rancher:
  services:
    rancher-agent1:
      image: rancher/agent:v0.8.2
      command: http://<rancher-server-ip>:8080/v1/projects/1a5/scripts/<registrationToken>
      privileged: true
      volumes:
        - /var/run/docker.sock:/var/run/docker.sock
      environment:
        CATTLE_HOST_LABELS: foo=bar
```
<br>

Adding more than one host label requires joining the additional host labels with an `&`

```yaml
#cloud-config
rancher:
  services:
    rancher-agent1:
    image: rancher/agent:v0.8.2
    command: http://<rancher-server-ip>:8080/v1/projects/1a5/scripts/<registrationToken>
    privileged: true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
    CATTLE_HOST_LABELS: foo=bar&hello=world
```

