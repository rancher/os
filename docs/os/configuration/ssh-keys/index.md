---
title: SSH Keys in RancherOS


---

## SSH Keys
---

RancherOS supports adding SSH keys through the [cloud-config]({{page.osbaseurl}}/configuration/#cloud-config) file. Within the cloud-config file, you simply add the ssh keys within the `ssh_authorized_keys` key.

```yaml
#cloud-config
ssh_authorized_keys:
  - ssh-rsa AAA...ZZZ example1@rancher
  - ssh-rsa BBB...ZZZ example2@rancher
```

When we pass the cloud-config file during the `ros install` command, it will allow these ssh keys to be associated with the **rancher** user. You can ssh into RancherOS using the key.

```
$ ssh -i /path/to/private/key rancher@<ip-address>
```

Please note that OpenSSH 7.0 and greater similarly disable the ssh-dss (DSA) public key algorithm. It too is weak and we recommend against its use.
