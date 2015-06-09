---
title: Installing to Disk
layout: default
---

## Installing RancherOS to Disk
---
RancherOS comes with a simple installer that will install RancherOS on a given target disk. To install RancherOS on a new disk, you can use the `rancheros-install` [command]({{site.baseurl}}/docs/rancheros-tools/rancheros-install). 




### Cloud Config

After you install RancherOS to disk, the rancher/rancher user/password will no longer be valid and you'll need to have added in SSH keys or another user within your [cloud config file]({{site.baseurl}}/docs/cloud-config/).

The easiest way to log in is to pass a `cloud-config.yml` file containing your public SSH keys. To learn more about what's supported in our cloud-config, please read this [doc]({{site.baseurl}}/docs/cloud-config/). 

The `rancheros-install` command will process your `cloud-config.yml` file specified with the `-c` flag. This file will also be placed onto the disk and installed to `/var/lib/rancher/conf/`. It will be evaluated on every boot and be converted to `/var/lib/rancher/conf/cloud-config-processed.yml`. 

By creating a cloud config file with a SSH key with no specific user, this allows you to SSH into the box. The yaml file would look like this:

```yaml
#cloud-config

ssh_authorized_keys:
- ssh-rsa AAA... user@host
```

You can generate a new SSH key for `cloud-config.yml` file by following this [article](https://help.github.com/articles/generating-ssh-keys/). 

Alternatively, if you have access to your local machine, you can copy your existing public SSH key into RancherOS before installing to disk. 

1. Check to see if you have SSH keys on your local computer. 

    ```bash
    $ ls -al ~/.ssh
    # Lists all the files in your .ssh directory, if they exist
    ```
2. You are looking to see if you have one of the following:

    ```bash
    id_dsa.pub
    id_ecdsa.pub
    id_ed25519.pub
    id_rsa.pub
    ```

3. After you've ensured that you have a public SSH key, we'll proceed with copyubg the public SSH key from our computer to our VM. In our example, the public key that we're using is `id_rsa.pub`.

    ```bash
    $ scp -r computer_username@computer_ip:~/.ssh/id_rsa.pub ./
    RSA key fingerprint is X:X:X:X:X.
    Are you sure you want to continue connecting (yes/no)? yes
    Warning: Permanently added 'computer_ip' (RSA) to the list of known hosts. 
    Password: 
    id_rsa.pub                              100%    422 0.4KB/s    00:00
    $ ls
    id_rsa.pub
    $ cat id_rsa.pub > cloud_config.yml
    $ vi cloud_config.yml
    ```

4. You'll need to edit the `cloud_config.yml` so that it matches the syntax of a cloud config file. Yaml files are very particular about their white spacing, so please note the spaces in our file!

Now that our `cloud_config.yml` contains our public SSH key, we can move on to installing RancherOS to disk!

### Using `rancheros-install` to Install RancherOS 

The `rancheros-install` command orchestrates the installation from the `rancher/os` container. You will need to have already created a cloud config file and found the target disk.

```bash
$ sudo rancheros-install -c cloud_config.yml -d /dev/sda 
All data will be wiped from this device
Partition: true
DEVICE: /dev/sda
Are you sure you want to continue? [yN]
```

You will be prompted to see if you want to continue. Type **y**.

```bash
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

### SSH into RancherOS

After installing RancherOS, you can ssh into RancherOS using your private key and the **rancher** user.

```bash
$ ssh -i /path/to/private/key rancher@<ip-address>
```

